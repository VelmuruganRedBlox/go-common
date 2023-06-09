package deeplink

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/http_client"
	"time"
)

type Service struct {
	httpClient *http_client.HttpClient
	ctx        context.Context
	config     Config
	url        string
}

func NewService(config Config, httpClient *http_client.HttpClient, ctx context.Context) IService {
	url := "https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=" + config.Key
	return &Service{
		config:     config,
		url:        url,
		httpClient: httpClient,
		ctx:        ctx,
	}
}

func (s *Service) generateDeeplink(link string) (string, error) {
	requestBody := firebaseCreateDeeplinkRequest{
		DynamicLinkInfo: dynamicLinkInfo{
			DomainUriPrefix: s.config.DomainURIPrefix,
			Link:            link,
			AndroidInfo: androidInfo{
				AndroidPackageName: s.config.AndroidPackageName,
			},
			IosInfo: iosInfo{
				IosBundleId:   s.config.IOSBundleId,
				IosAppStoreId: s.config.IOSAppStoreId,
			},
		},
	}
	requestJsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	resp, err := s.httpClient.NewRequest(s.ctx).SetContentType("application/json").SetBody(requestJsonBody).Post(s.url)

	if err != nil {
		return "", err
	}

	var firebaseResp firebaseCreateDeeplinkResponse

	if err := json.Unmarshal(resp.Bytes(), &firebaseResp); err != nil {
		return "", err
	}
	return firebaseResp.ShortLink, nil
}

func (s *Service) GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error) {
	link := fmt.Sprintf("%v/%v/%v", s.config.URI, s.getShareType(contentType), contentId)
	shareCode := fmt.Sprintf("%v%v%v", contentId, userId, time.Now().Unix())
	link += fmt.Sprintf("?sharerId=%v&referredByType=shared_content&shareCode=%v", userId, shareCode)
	if len(referralCode) > 0 {
		link += fmt.Sprintf("&referralCode=%v", referralCode)
	}
	return s.generateDeeplink(link)
}

func (s *Service) getShareType(contentType eventsourcing.ContentType) string {
	switch contentType {
	case eventsourcing.ContentTypeMusic:
		return "music"
	case eventsourcing.ContentTypeSpot:
		return "spot"
	default:
		return "video"
	}
}
