package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lastforeverzl/barkme/mydb"
)

type Env struct {
	db mydb.Datastore
}

type dbFunc func(chan *mydb.UserChan, string, mydb.User)

func main() {
	db, err := mydb.NewDB("./config.json")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("Successfully connected!")
	db.InitSchema()
	env := &Env{db}

	hub := newHub(env)
	go hub.run()

	router := gin.Default()
	router.LoadHTMLFiles("test.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.html", nil)
	})
	// TODO: Create REST API
	router.POST("/user", env.newUserHandler)
	router.GET("/users/", env.allUsersHandler)
	router.PUT("/user/:id", env.updateUserHandler)
	router.PUT("/add-fav-user-to/:id", env.addFavUserHandler)
	router.PUT("/rm-fav-user-from/:id", env.rmFavUserHandler)

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(hub, c.Writer, c.Request)
	})

	router.Run(":8080")
}

func (env *Env) newUserHandler(c *gin.Context) {
	channel := make(chan *mydb.UserChan)
	result := make(chan gin.H)
	go env.db.CreateUser(channel)
	go func() {
		newUser := <-channel
		if newUser.Err != nil {
			result <- gin.H{
				"status": http.StatusBadRequest,
				"error":  newUser.Err,
			}
		}
		result <- gin.H{
			"status": http.StatusOK,
			"user":   newUser.User,
		}
	}()
	c.JSON(http.StatusOK, <-result)
	close(result)
}

func (env *Env) allUsersHandler(c *gin.Context) {
	channel := make(chan *mydb.AllUsers)
	result := make(chan gin.H)
	go env.db.GetAllUsers(channel)
	go func() {
		users := <-channel
		if users.Err != nil {
			result <- gin.H{
				"status": http.StatusBadRequest,
				"error":  users.Err,
			}
		}
		result <- gin.H{
			"status": http.StatusOK,
			"users":  users.Users,
		}
	}()
	c.JSON(http.StatusOK, <-result)
	close(result)
}

func modifyUserInDb(c *gin.Context, fn dbFunc) {
	id := c.Param("id")
	user := mydb.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  err.Error(),
		})
	} else {
		channel := make(chan *mydb.UserChan)
		result := make(chan gin.H)
		go fn(channel, id, user)
		go func() {
			user := <-channel
			if user.Err != nil {
				result <- gin.H{
					"status": http.StatusBadRequest,
					"error":  user.Err,
				}
			}
			result <- gin.H{
				"status": http.StatusOK,
				"user":   user.User,
			}
		}()
		c.JSON(http.StatusOK, <-result)
		close(result)
	}
}

func (env *Env) updateUserHandler(c *gin.Context) {
	modifyUserInDb(c, env.db.UpdateUser)
}

func (env *Env) addFavUserHandler(c *gin.Context) {
	modifyUserInDb(c, env.db.AddFavUser)
}

func (env *Env) rmFavUserHandler(c *gin.Context) {
	modifyUserInDb(c, env.db.RemoveFavUser)
}
