package sdkInit

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func (t *Application) BurnImageNFT(chaincodeID string, args []string) (string, error) {
	var tempArgs [][]byte
	for i := 1; i < len(args); i++ {
		tempArgs = append(tempArgs, []byte(args[i]))
	}

	request := channel.Request{ChaincodeID: chaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}}
	response, err := t.SdkEnvInfo.ChClient.Execute(request)
	if err != nil {
		// 资产转移失败
		return "", err
	}

	//fmt.Println("============== response:",response)

	return string(response.TransactionID), nil
}
