package user

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserWrapperMock struct {
	GetUsersFn       func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
	GetUsersDetailFn func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
	GetProfileBulkFn func(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
}

func (m *UserWrapperMock) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	return m.GetUsersFn(userIds, apmTransaction, forceLog)
}

func (m *UserWrapperMock) GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan {
	return m.GetUsersDetailFn(userIds, apmTransaction, forceLog)
}

func (m *UserWrapperMock) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	return m.GetProfileBulkFn(currentUserId, userIds, apmTransaction, forceLog)
}

func GetMock() IUserWrapper { // for compiler errors
	return &UserWrapperMock{}
}
