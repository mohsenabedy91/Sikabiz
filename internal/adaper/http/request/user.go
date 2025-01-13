package request

type UserIDUri struct {
	ID uint64 `uri:"userID" binding:"required,number" example:"1"`
}
