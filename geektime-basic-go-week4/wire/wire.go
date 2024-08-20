//go:build wireinject

package wire

import (
	"gitee.com/geekbang/basic-go/wire/repository"
	"gitee.com/geekbang/basic-go/wire/repository/dao"
	"github.com/google/wire"
)

func InitUserRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, InitDB, dao.NewUserDAO)
	return &repository.UserRepository{}
}
