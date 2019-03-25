package azureblob

import (
	"testing"

	"io/ioutil"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
)

var activityMetadata *activity.Metadata

func getActivityMetadata() *activity.Metadata {

	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}

		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}

	return activityMetadata
}

func TestCreate(t *testing.T) {

	act := NewActivity(getActivityMetadata())

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
}

func TestRun(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	tc.SetSetting(AZURE_STORAGE_ACCOUNT, "flogo")
	tc.SetSetting(AZURE_STORAGE_ACCESS_KEY, "")
	tc.SetSetting(Method, "upload")
	tc.SetSetting(ContainerName, "sample")

	tc.SetInput("file", "abc.txt")
	tc.SetInput("data", 2)

	done, _ := act.Eval(tc)

	if !done {
		t.Error("activity should be done")
		return
	}

}
