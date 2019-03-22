package azureblob

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

var log = logger.GetLogger("activity-tibco-azureblob")

// AWSSNS Structure for the AWSNS activity
type AZBLOB struct {
	metadata *activity.Metadata
	settings Settings
	log      logger.Logger
}
type Settings struct {
	AZURE_STORAGE_ACCOUNT    string `md:"azure_storage_account,required"`
	AZURE_STORAGE_ACCESS_KEY string `md:"azure_storage_access_key,required"`
	Method                   string `md:"method,required"`
	ContainerName            string `md:"container_name,required"`
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &AZBLOB{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *AZBLOB) Metadata() *activity.Metadata {
	return a.metadata
}

func handleErrors(err error, log logger.Logger) error {
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
			switch serr.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:

				return errors.New("Received 409. Container already exists")
			}
		}
		log.Info(err)
	}

	return nil
}

func (a *AZBLOB) Eval(ctx activity.Context) (done bool, err error) {

	accountName, accountKey := a.settings.AZURE_STORAGE_ACCOUNT, a.settings.AZURE_STORAGE_ACCESS_KEY

	inputFile := ctx.GetInput("file").(string)

	inputData := ctx.GetInput("data").(string)

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		a.log.Debugf("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	containerName := a.settings.ContainerName

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	containerURL := azblob.NewContainerURL(*URL, p)
	bctx := context.Background()

	a.log.Infof("Executing method ", a.settings.Method)
	switch a.settings.Method {

	case "upload":
		a.log.Infof("Creating a container named %s\n", containerName)

		// This example uses a never-expiring context
		_, err = containerURL.Create(bctx, azblob.Metadata{}, azblob.PublicAccessNone)
		if err != nil {
			return true, err
		}
		err = handleErrors(err, a.log)

		if err != nil {
			return true, err
		}
		a.log.Info("Creating a dummy file to test the upload and download\n")
		err = ioutil.WriteFile(inputFile, []byte(inputData), 0700)
		err = handleErrors(err, a.log)

		if err != nil {
			return true, err
		}

		// Here's how to upload a blob.
		blobURL := containerURL.NewBlockBlobURL(inputFile)
		file, err := os.Open(inputFile)
		err = handleErrors(err, a.log)

		if err != nil {
			return true, err
		}

		a.log.Infof("Uploading the file with blob name: %s\n", inputFile)
		_, err = azblob.UploadFileToBlockBlob(bctx, file, blobURL, azblob.UploadToBlockBlobOptions{
			BlockSize:   4 * 1024 * 1024,
			Parallelism: 16})

	case "list":
		out := make(map[string]interface{})
		a.log.Info("Listing the blobs in the container:")
		for marker := (azblob.Marker{}); marker.NotDone(); {
			// Get a result segment starting with the blob indicated by the current Marker.
			listBlob, err := containerURL.ListBlobsFlatSegment(bctx, marker, azblob.ListBlobsSegmentOptions{})
			err = handleErrors(err, a.log)

			if err != nil {
				return true, err
			}

			// ListBlobs returns the start of the next segment; you MUST use this to get
			// the next segment (after processing the current result segment).
			marker = listBlob.NextMarker

			// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
			for _, blobInfo := range listBlob.Segment.BlobItems {
				a.log.Infof(" Blob name: " + blobInfo.Name + "\n")
				out[blobInfo.Name] = blobInfo

			}
		}
		ctx.SetOutput("result", out)

	}
	return true, nil
}
