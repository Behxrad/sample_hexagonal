package errcodes

import "192.168.100.6/coreinno/common.git/errs"

const (
	// 4xx...x range
	ErrInvalidProductCount = errs.Code(40001)
	ErrProductExists       = errs.Code(40002)

	// 5xx...x range
	ErrDatabaseConnectivityIssue = errs.Code(50001)
)
