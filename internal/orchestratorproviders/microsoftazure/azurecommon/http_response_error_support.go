package azurecommon

import (
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func ParseResponseError(err error) error {
	if err == nil {
		return nil
	}

	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		//log.Error("error response", "Error Details", azureResponseErrorDetails(respErr))
		//return fmt.Errorf("error response httpStatus=%d ErrorCode=%s", respErr.StatusCode, respErr.ErrorCode)
		return toError(respErr)
	}
	//log.Error("error response", "Error Details", err)
	return err
}

func toError(respErr *azcore.ResponseError) error {
	if respErr.RawResponse == nil || respErr.RawResponse.Request == nil {
		return fmt.Errorf("error response httpStatus=%d ErrorCode=%s", respErr.StatusCode, respErr.ErrorCode)
	}

	return fmt.Errorf("error response %s", respErr.Error())
}
