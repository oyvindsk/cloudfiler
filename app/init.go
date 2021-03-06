package app

import (
	"github.com/netbrain/cloudfiler/app/controller"
	"github.com/netbrain/cloudfiler/app/repository/fs"
	"github.com/netbrain/cloudfiler/app/web"
	"github.com/netbrain/cloudfiler/app/web/auth"
	"github.com/netbrain/cloudfiler/app/web/handler"
	"github.com/netbrain/cloudfiler/app/web/interceptor"
	"log"
)

var muxer = web.NewMuxer()
var authenticator auth.Authenticator
var WebHandler = new(interceptor.InterceptorChain)

func init() {
	initApplication()
	initRoutes()
}

func initApplication() {
	log.Println("Initializing application dependencies...")

	//init repositories
	userRepo := fs.NewUserRepository()
	roleRepo := fs.NewRoleRepository(userRepo)
	fileRepo := fs.NewFileRepository(userRepo, roleRepo)

	//init interceptor chain
	authenticator = auth.NewAuthenticator(userRepo, roleRepo, "/auth/login", "/")
	WebHandler.AddInterceptor(authenticator)
	WebHandler.AddInterceptor(muxer)

	//init controllers
	userController := controller.NewUserController(userRepo)
	roleController := controller.NewRoleController(roleRepo, userRepo)
	fileController := controller.NewFileController(fileRepo)

	//init handlers
	indexHandler := handler.NewIndexHandler(authenticator, userController)
	initHandler := handler.NewInitHandler(userController, roleController)
	userHandler := handler.NewUserHandler(userController)
	roleHandler := handler.NewRoleHandler(roleController, userController)
	authHandler := handler.NewAuthHandler(authenticator)
	fileHandler := handler.NewFileHandler(
		authenticator,
		fileController,
		userController,
		roleController,
	)

	//wire it all up
	log.Println("Adding web handlers...")
	muxer.AddHandler(indexHandler)
	muxer.AddHandler(initHandler)
	muxer.AddHandler(authHandler)
	muxer.AddHandler(userHandler)
	muxer.AddHandler(roleHandler)
	muxer.AddHandler(fileHandler)
}

func initRoutes() {
	log.Println("Adding routing table...")

	//Index
	addRoute(handler.IndexHandler.Index, "/")

	//Init
	addRoute(handler.InitHandler.Init, "/init")

	//Auth
	addRoute(handler.AuthHandler.Login, "/auth/login")
	addRoute(handler.AuthHandler.Logout, "/auth/logout")

	//User
	addRoute(handler.UserHandler.List, "/user/list", "Admin")
	addRoute(handler.UserHandler.Create, "/user/create", "Admin")
	addRoute(handler.UserHandler.Retrieve, "/user/retrieve", "Admin")
	addRoute(handler.UserHandler.Update, "/user/update", "Admin")

	//Role
	addRoute(handler.RoleHandler.List, "/role/list", "Admin")
	addRoute(handler.RoleHandler.Create, "/role/create", "Admin")
	addRoute(handler.RoleHandler.Retrieve, "/role/retrieve", "Admin")
	addRoute(handler.RoleHandler.Update, "/role/update", "Admin")
	addRoute(handler.RoleHandler.Delete, "/role/delete", "Admin")
	addRoute(handler.RoleHandler.RemoveUser, "/role/users/remove", "Admin")
	addRoute(handler.RoleHandler.AddUser, "/role/users/add", "Admin")

	//File
	addRoute(handler.FileHandler.List, "/file/list")
	addRoute(handler.FileHandler.Upload, "/file/upload")
	addRoute(handler.FileHandler.Retrieve, "/file/retrieve")
	addRoute(handler.FileHandler.Download, "/file/download")
	addRoute(handler.FileHandler.AddUsers, "/file/users/add")
	addRoute(handler.FileHandler.RemoveUsers, "/file/users/remove")
	addRoute(handler.FileHandler.AddRoles, "/file/roles/add")
	addRoute(handler.FileHandler.RemoveRoles, "/file/roles/remove")
	addRoute(handler.FileHandler.Tags, "/file/tags")
	addRoute(handler.FileHandler.AddTags, "/file/tags/add")
	addRoute(handler.FileHandler.RemoveTags, "/file/tags/remove")
	addRoute(handler.FileHandler.SetTags, "/file/tags/set")
	addRoute(handler.FileHandler.Delete, "/file/delete")
	addRoute(handler.FileHandler.Search, "/file/search")
	addRoute(handler.FileHandler.Update, "/file/update")
}

func addRoute(handler interface{}, path string, requiredRoles ...string) {
	log.Printf("Adding route '%s' with required roles: %v", path, requiredRoles)
	muxer.AddAction(path, handler)

	if len(requiredRoles) > 0 {
		authenticator.SetRequiredPrivileges(path, requiredRoles...)
	}
}
