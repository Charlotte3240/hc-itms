package services

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image/png"
	"io"
)

// convertCgbiToStandardPNG strips the CgBI chunk and fixes IDAT data
// to produce a standard PNG that browsers can render correctly.
func convertCgbiToStandardPNG(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	// Check PNG signature
	sig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	for i := 0; i < 8; i++ {
		if data[i] != sig[i] {
			return nil, fmt.Errorf("not a PNG file")
		}
	}

	// Parse all chunks
	chunks, err := parsePNGChunks(data)
	if err != nil {
		return nil, err
	}

	// Check if CgBI chunk exists
	hasCgBI := false
	cgbiType := [4]byte{'C', 'g', 'B', 'I'}
	for _, c := range chunks {
		if c.chunkType == cgbiType {
			hasCgBI = true
			break
		}
	}

	if !hasCgBI {
		return data, nil
	}

	// Concatenate all IDAT chunk data (PNG spec: multiple IDAT chunks form a single data stream)
	idatType := [4]byte{'I', 'D', 'A', 'T'}
	var idatData []byte
	for _, c := range chunks {
		if c.chunkType == idatType {
			idatData = append(idatData, c.data...)
		}
	}

	// Decompress the concatenated IDAT data
	decompressed, err := decompressCgbiIDAT(idatData)
	if err != nil {
		return data, fmt.Errorf("decompress IDAT: %w", err)
	}

	// Recompress with standard deflate
	var compressed bytes.Buffer
	w, _ := flate.NewWriter(&compressed, flate.DefaultCompression)
	w.Write(decompressed)
	w.Close()

	// Rebuild PNG without CgBI chunk, with new IDAT data
	var result []byte
	result = append(result, sig...)

	idatWritten := false
	iendType := [4]byte{'I', 'E', 'N', 'D'}

	for _, c := range chunks {
		// Skip CgBI chunk
		if c.chunkType == cgbiType {
			continue
		}

		// Replace first IDAT chunk with recompressed data, skip rest
		if c.chunkType == idatType {
			if !idatWritten {
				result = appendChunk(result, c.chunkType, compressed.Bytes())
				idatWritten = true
			}
			continue
		}

		// Keep other chunks as-is
		result = appendChunk(result, c.chunkType, c.data)

		if c.chunkType == iendType {
			break
		}
	}

	// Validate the result
	if _, err := png.Decode(bytes.NewReader(result)); err != nil {
		return data, nil
	}

	return result, nil
}

type pngChunk struct {
	chunkType [4]byte
	data      []byte
	crc       uint32
}

func parsePNGChunks(data []byte) ([]pngChunk, error) {
	var chunks []pngChunk
	pos := 8 // skip signature

	for pos < len(data) {
		if pos+8 > len(data) {
			break
		}

		length := binary.BigEndian.Uint32(data[pos : pos+4])
		pos += 4

		var chunkType [4]byte
		copy(chunkType[:], data[pos:pos+4])
		pos += 4

		if pos+int(length)+4 > len(data) {
			break
		}

		chunkData := make([]byte, length)
		copy(chunkData, data[pos:pos+int(length)])
		pos += int(length)

		crc := binary.BigEndian.Uint32(data[pos : pos+4])
		pos += 4

		chunks = append(chunks, pngChunk{
			chunkType: chunkType,
			data:      chunkData,
			crc:       crc,
		})

		// Stop at IEND
		if string(chunkType[:]) == "IEND" {
			break
		}
	}

	return chunks, nil
}

func appendChunk(data []byte, chunkType [4]byte, chunkData []byte) []byte {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(chunkData)))

	data = append(data, length...)
	data = append(data, chunkType[:]...)
	data = append(data, chunkData...)

	// Calculate CRC
	crcData := make([]byte, 0, 4+len(chunkData))
	crcData = append(crcData, chunkType[:]...)
	crcData = append(crcData, chunkData...)
	crc := crc32.ChecksumIEEE(crcData)

	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)
	data = append(data, crcBytes...)

	return data
}

// decompressCgbiIDAT decompresses CgBI-format IDAT data
func decompressCgbiIDAT(data []byte) ([]byte, error) {
	// CgBI uses raw deflate without zlib wrapper.
	// Try raw deflate first, then try stripping zlib header (2 bytes) if that fails.

	// Attempt 1: raw deflate
	reader := flate.NewReader(bytes.NewReader(data))
	result, err := io.ReadAll(reader)
	reader.Close()
	if err == nil && len(result) > 0 {
		return result, nil
	}

	// Attempt 2: zlib-wrapped (strip 2-byte header, 4-byte checksum)
	if len(data) > 6 {
		reader = flate.NewReader(bytes.NewReader(data[2:]))
		result, err = io.ReadAll(reader)
		reader.Close()
		if err == nil && len(result) > 0 {
			return result, nil
		}
	}

	return nil, fmt.Errorf("failed to decompress CgBI IDAT data")
}
