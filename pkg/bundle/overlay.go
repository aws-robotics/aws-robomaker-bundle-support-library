// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

// overlays and overlay mirror the json structure of the overlays documented in V2 format here:
// https://github.com/colcon/colcon-bundle/blob/master/BUNDLE_FORMAT.md
type overlays struct {
	Overlays []overlay `json:"overlays"`
}

type overlay struct {
	FileName string `json:"name"`
	Sha256   string `json:"sha256"`
	Offset   int    `json:"offset"`
	Size     int    `json:"size"`
}
