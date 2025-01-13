package request

type UserIDUri struct {
	ID uint64 `uri:"ID" binding:"required,number" example:"1"`
}
