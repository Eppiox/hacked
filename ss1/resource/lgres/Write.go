package lgres

import (
	"fmt"
	"io"

	"github.com/inkyblackness/hacked/ss1/resource"
)

// Write serializes the resources from given provider into the target.
// It is a convenience function for using Writer.
func Write(target io.WriteSeeker, source resource.Provider) error {
	writer, writerErr := NewWriter(target)
	if writerErr != nil {
		return writerErr
	}

	for _, id := range source.IDs() {
		entry, resourceErr := source.Resource(id)
		if resourceErr != nil {
			return resourceErr
		}

		if entry.Compound {
			resourceWriter, resourceWriterErr := writer.CreateCompoundResource(id, entry.ContentType, entry.Compressed)
			if resourceWriterErr != nil {
				return resourceWriterErr
			}
			copyErr := copyBlocks(entry, func() io.Writer { return resourceWriter.CreateBlock() })
			if copyErr != nil {
				return copyErr
			}
		} else if entry.BlockCount() == 1 {
			blockWriter, resourceWriterErr := writer.CreateResource(id, entry.ContentType, entry.Compressed)
			if resourceWriterErr != nil {
				return resourceWriterErr
			}
			copyErr := copyBlocks(entry, func() io.Writer { return blockWriter })
			if copyErr != nil {
				return copyErr
			}
		} else {
			return fmt.Errorf("simple resource %v has wrong number of blocks", id)
		}
	}

	return writer.Finish()
}

func copyBlocks(source resource.BlockProvider, nextWriter func() io.Writer) error {
	for blockIndex := 0; blockIndex < source.BlockCount(); blockIndex++ {
		blockReader, blockErr := source.Block(blockIndex)
		if blockErr != nil {
			return blockErr
		}
		_, copyErr := io.Copy(nextWriter(), blockReader)
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}
