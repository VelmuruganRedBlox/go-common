package s3

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUseContentUploaderProxy(t *testing.T) {
	finalUrl := UseContentUploaderProxy(context.TODO(),
		"https://lit-video-cdn-dta-dev.s3.eu-west-1.amazonaws.com/user/112125/8e63485b-840c-4636-a858-7abc6c6a69cd.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAS3VHOT3G43A3KO5Q%2F20220704%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Date=20220704T181102Z&X-Amz-Expires=600&X-Amz-Security-Token=FwoGZXIvYXdzEHwaDDRqFqJRiOnZUUIA%2FyKEBIQAxoW45d3FmaHcfvyH1OXvQckXO22fgyzMCDHRrkVddDVwNc3lOJfdNWZ1WWx3OV396DUGT%2FBFlt8VME7Cdvfqqfm2CEaFPFDKmHNheJDUsbI7B3PPhqgp7T0tT0uqTtSPXRqD741vT5OtklioMVpHprMJ16Uk47nr%2BqBADNtWQCPy0RO%2BKlN4gsPU8jhzZ3s%2FJbZbghzm19O4GtiWkYHHbDMELqyyzAvo98Y%2B%2FHGlQEvGElI6Xj5l1kt8XQ3DKuBiX56FBGmrlIzzyp0A%2Fkmm2ttUQWFECWYBEfzbMZvIWrv67R4k5hWNa2yZBeFEZ0nTHWm2CcYYuVMk9QqRsQVJ%2BJWPTuRO%2FH9vNwiBHvhedzkzeC3M1Knx%2F8v9wdsm1xP2bHBjEbQ6iRuHgJdjxmAak3kKaEy5k8j%2FH%2BVZBNQMKjbPp75%2B6r%2FYQ0saG918ZJqo%2FmuxLzxWN0PYznPR%2ByQFm44nG%2FylEks%2B4VJyEBdwdCp%2BS65tnGkuvGYC3eaZvEfmrVwd1%2FnfT%2FX6Ua6WxhDde%2Fu12p7XjnHs8K2IaVUg52fcfS9Xs%2FOxn%2B8uPZqmvXONdzUI3dO2lx1Ic8iuY3vCxKqcyBKqj%2Fxrrwi5fQ4t1pCpbRJl9rOxamIILGKDnQbgBlONanCXol34ZJbv8Mhg1keTlX8lTfAQyCQ3gO%2FxPDALMSjk2IyWBjIqC%2BbrYrkcmn5DG1EDF%2FzEO6QkEFuk1Ih%2FBuJTqPfk%2BgLBGBZMUQW1Dbf7&X-Amz-SignedHeaders=host&X-Amz-Signature=f56d7ab6430c8deb9618dd70fa79b49b4b889ca5619960f491744bc38ee23bf5",
		"http://localhost:9092")

	assert.Equal(t, "http://localhost:9092/user/112125/8e63485b-840c-4636-a858-7abc6c6a69cd.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAS3VHOT3G43A3KO5Q%2F20220704%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Date=20220704T181102Z&X-Amz-Expires=600&X-Amz-Security-Token=FwoGZXIvYXdzEHwaDDRqFqJRiOnZUUIA%2FyKEBIQAxoW45d3FmaHcfvyH1OXvQckXO22fgyzMCDHRrkVddDVwNc3lOJfdNWZ1WWx3OV396DUGT%2FBFlt8VME7Cdvfqqfm2CEaFPFDKmHNheJDUsbI7B3PPhqgp7T0tT0uqTtSPXRqD741vT5OtklioMVpHprMJ16Uk47nr%2BqBADNtWQCPy0RO%2BKlN4gsPU8jhzZ3s%2FJbZbghzm19O4GtiWkYHHbDMELqyyzAvo98Y%2B%2FHGlQEvGElI6Xj5l1kt8XQ3DKuBiX56FBGmrlIzzyp0A%2Fkmm2ttUQWFECWYBEfzbMZvIWrv67R4k5hWNa2yZBeFEZ0nTHWm2CcYYuVMk9QqRsQVJ%2BJWPTuRO%2FH9vNwiBHvhedzkzeC3M1Knx%2F8v9wdsm1xP2bHBjEbQ6iRuHgJdjxmAak3kKaEy5k8j%2FH%2BVZBNQMKjbPp75%2B6r%2FYQ0saG918ZJqo%2FmuxLzxWN0PYznPR%2ByQFm44nG%2FylEks%2B4VJyEBdwdCp%2BS65tnGkuvGYC3eaZvEfmrVwd1%2FnfT%2FX6Ua6WxhDde%2Fu12p7XjnHs8K2IaVUg52fcfS9Xs%2FOxn%2B8uPZqmvXONdzUI3dO2lx1Ic8iuY3vCxKqcyBKqj%2Fxrrwi5fQ4t1pCpbRJl9rOxamIILGKDnQbgBlONanCXol34ZJbv8Mhg1keTlX8lTfAQyCQ3gO%2FxPDALMSjk2IyWBjIqC%2BbrYrkcmn5DG1EDF%2FzEO6QkEFuk1Ih%2FBuJTqPfk%2BgLBGBZMUQW1Dbf7&X-Amz-SignedHeaders=host&X-Amz-Signature=f56d7ab6430c8deb9618dd70fa79b49b4b889ca5619960f491744bc38ee23bf5",
		finalUrl)
}
