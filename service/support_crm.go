package service

import (
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/tel4vn/fins-microservices/common/log"
	"github.com/tel4vn/fins-microservices/model"
)

func SendMessageToAdminCrm(ctx context.Context, event any, authUser *model.AuthUser) {
	url := API_CRM + "/v1/crm/user-crm?level=user&unit_uuid=" + authUser.UnitUuid
	client := resty.New()
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+authUser.Token).
		Get(url)

	if err != nil {
		log.Error(err)
		return
	}
	if res.StatusCode() == 200 {
		var responseData model.ResponseData
		err = json.Unmarshal(res.Body(), &responseData)
		if err != nil {
			log.Error(err)
			return
		}

		userUuids := []string{}

		for _, item := range responseData.Data {
			userUuid, ok := item["user_uuid"].(string)
			if !ok {
				log.Error("user_uuid not found or not a string")
				continue
			}
			userUuids = append(userUuids, userUuid)
		}
		if len(userUuids) > 0 {
			if err := PublishMessageToMany(userUuids, event); err != nil {
				log.Error(err)
				return
			}
		}
	}
}
