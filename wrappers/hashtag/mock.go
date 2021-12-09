package hashtag

import (
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

type HashtagWrapperMock struct {
	GetHashtagsInternalFn func(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan
}

func (w *HashtagWrapperMock) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan {
	return w.GetHashtagsInternalFn(hashtags, omitHashtags, limit, offset, withViews, apmTransaction, forceLog)
}

func GetMock() IHashtagWrapper {
	return &HashtagWrapperMock{}
}
