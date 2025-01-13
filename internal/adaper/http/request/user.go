package request

type UserUUIDUri struct {
	UUIDStr string `uri:"userID" binding:"required,uuid" example:"8f4a1582-6a67-4d85-950b-2d17049c7385"`
}
