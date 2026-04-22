package domain

type ProductRepository interface {
	GetAll() ([]Product, error)
	GetByID(id int64) (*Product, error)
}

type UserRepository interface {
	FindByFirebaseUID(uid string) (*User, error)
	Create(user *User) error
}
