package binman

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rjbrown57/binman/pkg/gh"
	log "github.com/rjbrown57/binman/pkg/logging"
)

type ReleaseStatusAction struct {
	r           *BinmanRelease
	releasePath string
}

func (r *BinmanRelease) AddReleaseStatusAction(releasePath string) Action {
	return &ReleaseStatusAction{
		r,
		releasePath,
	}
}

// ReleaseStatusAction verifies whether we have work to do
func (action *ReleaseStatusAction) execute() error {

	action.r.setPublisPath(action.releasePath, *action.r.githubData.TagName)
	_, err := os.Stat(action.r.publishPath)

	// If err nil we already have this version, send custom error so gosyncrepo knows to end actions
	// Default to capture any other error cases
	switch err {
	case nil:
		//log.Infof("Latest version is %s %s is up to date", *action.r.githubData.TagName, action.r.Repo)
		return fmt.Errorf("%s", "Noupdate")
	default:
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
}

type SetUrlAction struct {
	r *BinmanRelease
}

func (r *BinmanRelease) AddSetUrlAction() Action {
	return &SetUrlAction{
		r,
	}
}

func (action *SetUrlAction) execute() error {
	// If user has set an external url use that to grab target
	if action.r.ExternalUrl != "" {
		log.Debugf("User specified url %s", action.r.dlUrl)
		action.r.dlUrl = formatString(action.r.ExternalUrl, action.r.getDataMap())
		action.r.assetName = filepath.Base(action.r.dlUrl)
		return nil
	}

	// If the user has requested a specifc asset check for that
	if action.r.ReleaseFileName != "" {
		rFilename := formatString(action.r.ReleaseFileName, action.r.getDataMap())
		log.Debugf("Get asset by name %s", rFilename)
		action.r.assetName, action.r.dlUrl = gh.GetAssetbyName(rFilename, action.r.githubData.Assets)
	} else {
		// Attempt to find the asset via arch/os
		log.Debugf("Attempt to find asset %s", action.r.ReleaseFileName)
		action.r.assetName, action.r.dlUrl = gh.FindAsset(action.r.Arch, action.r.Os, action.r.Version, action.r.project, action.r.githubData.Assets)
	}

	// If at this point dlUrl is not set we have an issue
	if action.r.dlUrl == "" {
		return fmt.Errorf("Target release asset not found for %s", action.r.Repo)
	}

	return nil
}

type SetArtifactPathAction struct {
	r           *BinmanRelease
	releasePath string
}

func (r *BinmanRelease) AddSetArtifactPathAction(releasePath string) Action {
	return &SetArtifactPathAction{
		r,
		releasePath,
	}
}

func (action *SetArtifactPathAction) execute() error {
	action.r.setArtifactPath(action.releasePath, action.r.assetName)
	err := CreateDirectory(action.r.publishPath)
	// At this point we have created something during the release process
	// so we set cleanupOnFailure to true in case we hit an issue further down the line
	action.r.cleanupOnFailure = true
	return err
}