package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/openshift/appliance/pkg/executer"
)

// Response is what Cincinnati sends us when querying for releases in a channel
type Response struct {
	Nodes []Release `json:"nodes"`
}

// Release describes a release payload
type Release struct {
	Version string `json:"version"`
	Payload string `json:"payload"`
}

// Release describes a generally available release payload
type OcpRelease struct {
	// Version is the minor version to search for
	Version string `json:"version"`
	// Channel is the release channel to search in
	Channel ReleaseChannel `json:"channel"`
	// Architecture is the architecture for the release.
	// Defaults to amd64.
	Architecture string `json:"architecture,omitempty"`
}

type ReleaseChannel string

const (
	ReleaseChannelStable    ReleaseChannel = "stable"
	ReleaseChannelFast      ReleaseChannel = "fast"
	ReleaseChannelCandidate ReleaseChannel = "candidate"
	ReleaseChannelEUS       ReleaseChannel = "eus"
)

// Graph is the interface for fetching info from api.openshift.com/api/upgrades_info/graph
type Graph interface {
	GetReleaseImage(version string, channel *string, arch string) (string, string, error)
}

type graph struct {
	executer executer.Executer
}

func NewGraph() Graph {
	return &graph{
		executer: executer.NewExecuter(),
	}
}

const (
	cincinnatiAddress = "https://api.openshift.com/api/upgrades_info/graph"
)

var (
	majorMinorRegExp = regexp.MustCompile(`^(?P<majorMinor>(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*))\.?.*`)
)

func (g *graph) GetReleaseImage(version string, channel *string, arch string) (string, string, error) {
	var releaseChannel ReleaseChannel
	if channel == nil {
		releaseChannel = ReleaseChannelStable
	} else {
		releaseChannel = ReleaseChannel(*channel)
	}

	release := OcpRelease{
		Version:      version,
		Channel:      releaseChannel,
		Architecture: arch,
	}
	client := retryablehttp.NewClient()
	client.Logger = nil
	payload, version, err := g.resolvePullSpec(client.StandardClient(), cincinnatiAddress, release)
	if err != nil {
		return "", "", err
	}

	return payload, version, nil
}

// Copied from ci-tools (https://github.com/openshift/ci-tools/blob/master/pkg/release/official/client.go)

func (g *graph) resolvePullSpec(client *http.Client, endpoint string, release OcpRelease) (string, string, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/json")
	query := req.URL.Query()
	explicitVersion, channel, err := g.processVersionChannel(release.Version, release.Channel)
	if err != nil {
		return "", "", err
	}
	targetName := "latest release"
	if !explicitVersion {
		targetName = release.Version
	}
	query.Add("channel", channel)
	query.Add("arch", string(release.Architecture))
	req.URL.RawQuery = query.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to request %s: %w", targetName, err)
	}
	if resp == nil {
		return "", "", fmt.Errorf("failed to request %s: got a nil response", targetName)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, readErr := io.Copy(&buf, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to request %s: server responded with %d: %s", targetName, resp.StatusCode, buf.String())
	}
	if readErr != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", readErr)
	}
	response := Response{}
	err = json.Unmarshal(buf.Bytes(), &response)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if len(response.Nodes) == 0 {
		return "", "", fmt.Errorf("failed to request %s from %s: server returned empty list of releases (despite status code 200)", targetName, req.URL.String())
	}

	if explicitVersion {
		for _, node := range response.Nodes {
			if node.Version == release.Version {
				return node.Payload, node.Version, nil
			}
		}
		return "", "", fmt.Errorf("failed to request %s from %s: version not found in list of releases", release.Version, req.URL.String())
	}

	pullspec, version := g.latestPullSpecAndVersion(response.Nodes)
	return pullspec, version, nil
}

// processVersionChannel takes the configured version and channel and
// returns:
//
//   - Whether the version is explicit (e.g. 4.7.0) or just a
//     major.minor (e.g. 4.7).
//   - The appropriate channel for a Cincinnati request, e.g. stable-4.7.
//   - Any errors that turn up while processing.
func (g *graph) processVersionChannel(version string, channel ReleaseChannel) (explicitVersion bool, cincinnatiChannel string, err error) {
	explicitVersion, majorMinor, err := g.extractMajorMinor(version)
	if err != nil {
		return false, "", err
	}
	if strings.HasSuffix(string(channel), fmt.Sprintf("-%s", majorMinor)) {
		return explicitVersion, string(channel), nil
	}

	return explicitVersion, fmt.Sprintf("%s-%s", channel, majorMinor), nil
}

// latestPullSpecAndVersion returns the pullSpec of the latest release in the list as a payload and version
func (g *graph) latestPullSpecAndVersion(options []Release) (string, string) {
	sort.Slice(options, func(i, j int) bool {
		vi := semver.MustParse(options[i].Version)
		vj := semver.MustParse(options[j].Version)
		return vi.GTE(vj) // greater, not less, so we get descending order
	})
	return options[0].Payload, options[0].Version
}

func (g *graph) extractMajorMinor(version string) (explicitVersion bool, majorMinor string, err error) {
	majorMinorMatch := majorMinorRegExp.FindStringSubmatch(version)
	if majorMinorMatch == nil {
		return false, "", fmt.Errorf("version %q does not begin with a major.minor version", version)
	}

	majorMinorIndex := majorMinorRegExp.SubexpIndex("majorMinor")
	majorMinor = majorMinorMatch[majorMinorIndex]
	explicitVersion = version != majorMinor

	return explicitVersion, majorMinor, nil
}
