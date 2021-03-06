package lgres

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/inkyblackness/hacked/ss1/resource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReaderFromReturnsErrorForNilSource(t *testing.T) {
	reader, err := ReaderFrom(nil)

	assert.Nil(t, reader, "reader should be nil")
	assert.Equal(t, errSourceNil, err)
}

func TestReaderFromReturnsInstanceOnEmptySource(t *testing.T) {
	source := bytes.NewReader(emptyResourceFile())
	reader, err := ReaderFrom(source)

	assert.Nil(t, err, "Error should be nil")
	assert.NotNil(t, reader)
}

func TestReaderFromReturnsErrorOnInvalidHeaderString(t *testing.T) {
	sourceData := emptyResourceFile()
	sourceData[10] = byte("A"[0])

	_, err := ReaderFrom(bytes.NewReader(sourceData))

	assert.Equal(t, errFormatMismatch, err)
}

func TestReaderFromReturnsErrorOnMissingCommentTerminator(t *testing.T) {
	sourceData := emptyResourceFile()
	sourceData[len(headerString)] = byte(0)

	_, err := ReaderFrom(bytes.NewReader(sourceData))

	assert.Equal(t, errFormatMismatch, err)
}

func TestReaderFromReturnsErrorOnInvalidDirectoryStart(t *testing.T) {
	sourceData := emptyResourceFile()
	sourceData[resourceDirectoryFileOffsetPos] = byte(0xFF)
	sourceData[resourceDirectoryFileOffsetPos+1] = byte(0xFF)

	_, err := ReaderFrom(bytes.NewReader(sourceData))

	assert.NotNil(t, err, "error expected")
}

func TestReaderFromCanDecodeExampleResourceFile(t *testing.T) {
	_, err := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	assert.Nil(t, err, "no error expected")
}

func TestReaderIDsReturnsTheStoredResourceIDsInOrderFromFile(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))

	assert.Equal(t, []resource.ID{exampleResourceIDSingleBlockResource, exampleResourceIDSingleBlockResourceCompressed,
		exampleResourceIDCompoundResource, exampleResourceIDCompoundResourceCompressed}, reader.IDs())
}

func TestReaderResourceReturnsErrorForUnknownID(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(emptyResourceFile()))
	resourceReader, err := reader.Resource(resource.ID(0x1111))
	assert.Nil(t, resourceReader, "no reader expected")
	assert.NotNil(t, err)
}

func TestReaderResourceReturnsAResourceReaderForKnownID(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	resourceReader, err := reader.Resource(exampleResourceIDSingleBlockResource)
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, resourceReader)
}

func TestReaderResourceReturnsResourceWithMetaInformation(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	info := func(resourceID resource.ID, name string, expected interface{}) string {
		return fmt.Sprintf("Resource 0x%04X should have %v = %v", resourceID.Value(), name, expected)
	}
	verifyResource := func(resourceID resource.ID, compound bool, contentType resource.ContentType, compressed bool) {
		resourceReader, _ := reader.Resource(resourceID)
		assert.Equal(t, compound, resourceReader.Compound, info(resourceID, "compound", compound))
		assert.Equal(t, contentType, resourceReader.ContentType, info(resourceID, "contentType", contentType))
		assert.Equal(t, compressed, resourceReader.Compressed, info(resourceID, "compressed", compressed))
	}
	verifyResource(exampleResourceIDSingleBlockResource, false, resource.ContentType(0x01), false)
	verifyResource(exampleResourceIDSingleBlockResourceCompressed, false, resource.ContentType(0x02), true)
	verifyResource(exampleResourceIDCompoundResource, true, resource.ContentType(0x03), false)
	verifyResource(exampleResourceIDCompoundResourceCompressed, true, resource.ContentType(0x04), true)
}

func TestReaderResourceReturnsSameInstance(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))

	c1, _ := reader.Resource(exampleResourceIDSingleBlockResource)
	c2, _ := reader.Resource(exampleResourceIDSingleBlockResource)

	assert.Equal(t, c1, c2, "Resources should be the same")
}

func TestReaderResourceWithUncompressedSingleBlockContent(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	resourceReader, _ := reader.Resource(exampleResourceIDSingleBlockResource)

	assert.Equal(t, 1, resourceReader.BlockCount())
	verifyBlockContent(t, resourceReader, 0, []byte{0x01, 0x01, 0x01})
}

func TestReaderResourceWithCompressedSingleBlockContent(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	resourceReader, _ := reader.Resource(exampleResourceIDSingleBlockResourceCompressed)

	assert.Equal(t, 1, resourceReader.BlockCount())
	verifyBlockContent(t, resourceReader, 0, []byte{0x02, 0x02})
}

func TestReaderResourceWithUncompressedCompoundContent(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	resourceReader, _ := reader.Resource(exampleResourceIDCompoundResource)

	assert.Equal(t, 2, resourceReader.BlockCount())
	verifyBlockContent(t, resourceReader, 0, []byte{0x30, 0x30, 0x30, 0x30})
	verifyBlockContent(t, resourceReader, 1, []byte{0x31, 0x31, 0x31})
}

func TestReaderResourceWithCompressedCompoundContent(t *testing.T) {
	reader, _ := ReaderFrom(bytes.NewReader(exampleResourceFile()))
	resourceReader, _ := reader.Resource(exampleResourceIDCompoundResourceCompressed)

	assert.Equal(t, 3, resourceReader.BlockCount())
	verifyBlockContent(t, resourceReader, 0, []byte{0x40, 0x40})
	verifyBlockContent(t, resourceReader, 1, []byte{0x41, 0x41, 0x41, 0x41})
	verifyBlockContent(t, resourceReader, 2, []byte{0x42})
}

func verifyBlockContent(t *testing.T, blockProvider resource.BlockProvider, blockIndex int, expected []byte) {
	blockReader, readerErr := blockProvider.Block(blockIndex)
	assert.Nil(t, readerErr, "error retrieving reader")
	require.NotNil(t, blockReader, "reader is nil")
	data, dataErr := ioutil.ReadAll(blockReader)
	assert.Nil(t, dataErr, "no error expected reading data")
	assert.Equal(t, expected, data)
}
