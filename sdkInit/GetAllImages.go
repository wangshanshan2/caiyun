package sdkInit

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (t *Application) GetAllImages(chaincodeID string) (string, error) {
	response, err := t.SdkEnvInfo.ChClient.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: "GetAllImages", Args: nil})
	if err != nil {
		return "", fmt.Errorf("failed to query: %v", err)
	}

	return string(response.Payload), nil
}
