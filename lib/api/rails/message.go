package rails

import "github.com/snowmerak/keycl/model/gen/rails"

func CommonResponse(success bool, message string) *rails.Message {
	return &rails.Message{
		Response: &rails.Message_CommonResponse{
			CommonResponse: &rails.CommonResponse{
				Success: success,
				Message: message,
			},
		},
	}
}

func ValueResponse(success bool, message string, value []byte) *rails.Message {
	return &rails.Message{
		Response: &rails.Message_ValueResponse{
			ValueResponse: &rails.ValueResponse{
				Success: success,
				Message: message,
				Value:   value,
			},
		},
	}
}
