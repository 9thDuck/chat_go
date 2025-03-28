package domain

type User struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	HashedPassword string `json:"-"`
	PublicKey      string `json:"publicKey"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ProfilePic     string `json:"profilePic"`
	RoleID         int64  `json:"roleId"`
	Role           *Role  `json:"role"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type Message struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"senderId"`
	ReceiverID  int64     `json:"receiverId"`
	Content     string    `json:"content"`
	Attachments *[]string `json:"attachments"`
	IsRead      bool      `json:"isRead"`
	IsDelivered bool      `json:"isDelivered"`
	Version     int64     `json:"version"`
	Edited      bool      `json:"edited"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type Contact struct {
	UserID    int64  `json:"user_id"`
	ContactID int64  `json:"contact_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ContactRequest struct {
	SenderID         int64  `json:"senderId"`
	SenderUsername   string `json:"senderUsername"`
	ReceiverID       int64  `json:"receiverId"`
	ReceiverUsername string `json:"receiverUsername"`
	CreatedAt        string `json:"createdAt"`
	MessageContent   string `json:"messageContent"`
}
