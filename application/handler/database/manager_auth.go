package database

import (
	"github.com/admpub/nging/application/library/dbmanager"
	"github.com/admpub/nging/application/library/dbmanager/driver"
	"github.com/admpub/nging/application/model"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

func authentication(mgr dbmanager.Manager, accountID uint, m *model.DbAccount) (err error, succeed bool) {
	ctx := mgr.Context()
	auth := mgr.Account()
	if accountID > 0 {
		auth.Driver = m.Engine
		auth.Username = m.User
		auth.Password = m.Password
		auth.Host = m.Host
		auth.Db = m.Name
		auth.AccountID = m.Id
		if len(m.Options) > 0 {
			options := echo.H{}
			com.JSONDecode(com.Str2bytes(m.Options), &options)
			auth.Charset = options.String(`charset`)
		}
		if len(auth.Charset) == 0 {
			auth.Charset = `utf8mb4`
		}
		key := driver.GenKey(``, ``, ``, ``, auth.AccountID)
		addAuth(mgr, auth, key)
		err = mgr.Run(auth.Driver, `login`)
		succeed = err == nil
		return
	}
	if accounts, exists := ctx.Session().Get(`dbAccounts`).(driver.AuthAccounts); exists {
		key := driver.GenKey(auth.Driver, auth.Username, auth.Host, auth.Db, accountID)
		data := accounts.Get(key)
		if data == nil {
			return
		}
		auth.CopyFrom(data)
		err = mgr.Run(auth.Driver, `login`)
		succeed = err == nil
		return
	}
	return
}
