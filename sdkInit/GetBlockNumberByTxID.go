package sdkInit

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (t *Application) GetBlockNumberByTxID(chaincodeID string, args []string) (string, error) {
	response, err := t.SdkEnvInfo.ChClient.Query(channel.Request{ChaincodeID: chaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}})
	if err != nil {
		return "", fmt.Errorf("failed to query: %v", err)
	}

	return string(response.Payload), nil
}
