// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package bundle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	v2MetadataFileName = "metadata.tar.gz"
	overlaysFileName   = "overlays.json"
)

func newBundleProcessorV2() bundleProcessor {
	return &bundleProcessorV2{}
}

// bundle v2 processor knows how to parse overlays and process them accordingly
type bundleProcessorV2 struct {
}

func (b *bundleProcessorV2) extract(inputStream io.ReadSeeker, bundleStore Cache) (Bundle, error) {

	// obtain the metadata from the bundle bytes
	metadataTarReader, metadataErr := getMetadataTarReader(inputStream)
	if metadataErr != nil {
		return nil, metadataErr
	}

	// get the list of overlays from the metadata
	overlays, overalysErr := getOverlays(metadataTarReader)
	if overalysErr != nil {
		return nil, overalysErr
	}

	var itemKeys []string

	// for every overlay, Extract them into the bundle store
	for _, overlay := range overlays.Overlays {

		fmt.Printf("Processing overlay: %+v\n", overlay)

		overlayReader, overlayErr := getReaderForOverlay(overlay, inputStream)
		if overlayErr != nil {
			return nil, overlayErr
		}

		tarGzExtractor := extractorFromFileName(overlayReader, overlay.FileName)
		if tarGzExtractor == nil {
			return nil, fmt.Errorf("cannot create extractor for overlay: %s", overlay.FileName)
		}

		// now, put into the bundle store, the store will take care of not extracting if it already exists
		_, putError := bundleStore.Put(overlay.Sha256, tarGzExtractor)
		if putError != nil {
			return nil, putError
		}
		itemKeys = append(itemKeys, overlay.Sha256)
	}

	//Seek to the end of the stream to expose completion to clients monitoring progress (we might not read everything)
	_, _ = inputStream.Seek(0, io.SeekEnd)

	// create a new bundle with item paths
	return newBundle(bundleStore, itemKeys), nil
}

// from the input stream get the metadata tar reader
func getMetadataTarReader(inputStream io.ReadSeeker) (*tar.Reader, error) {
	tarReader := tarReaderFromStream(inputStream)
	// skip past the version file and get to the metadata.tar.gz file
	tarReader.Next()
	metadataHeader, metadataErr := tarReader.Next()
	if metadataErr != nil {
		return nil, metadataErr
	}

	// ensure that we are now pointing to the metadata file
	if metadataHeader.Name != v2MetadataFileName {
		return nil, fmt.Errorf("unexpected metadata file: %s", metadataHeader.Name)
	}

	// create a limit reader in order to Extract this metadata file
	metadataReader := io.LimitReader(tarReader, metadataHeader.Size)

	// now, get a tar reader from this metadataReader
	// we know that this is a .tar.gz file
	metadataTarGzReader, gzErr := gzip.NewReader(metadataReader)
	if gzErr != nil {
		return nil, gzErr
	}
	// transform it into a tarReader
	return tar.NewReader(metadataTarGzReader), nil
}

func getOverlays(metadataTarReader *tar.Reader) (*overlays, error) {

	// iterate headers in the metadata tar file and process each file in the tar
	for {
		header, err := metadataTarReader.Next()
		if err == io.EOF {
			// we there are no more headers, we finish
			break
		} else if err != nil {
			return nil, err
		}

		// if we find the overlays file, read the bytes and parse it
		if header.Name == overlaysFileName {
			overlayBytes, overlayBytesErr := ioutil.ReadAll(metadataTarReader)
			if overlayBytesErr != nil {
				return nil, overlayBytesErr
			}

			var overlays overlays
			// unmarshal json
			jsonErr := json.Unmarshal(overlayBytes, &overlays)
			if jsonErr != nil {
				return nil, fmt.Errorf("unable to parse JSON of the overlays file: %s", jsonErr)
			}
			return &overlays, nil
		}
	}
	return nil, fmt.Errorf("overlays file not find in metadata")
}

func getReaderForOverlay(overlay overlay, inputStream io.ReadSeeker) (io.Reader, error) {
	// now we seek and create a limit reader, and get extractor
	_, seekError := inputStream.Seek(int64(overlay.Offset), io.SeekStart)

	if seekError != nil {
		return nil, fmt.Errorf("seekError: %v for %s", seekError, overlay.FileName)
	}

	// create a limit reader to read part of a file
	return io.LimitReader(inputStream, int64(overlay.Size)), nil
}
