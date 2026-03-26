package service

type Services struct {
	User    UserService
	Product ProductService
	Order   OrderService
	Token   TokenService
}

func NewServices(
	user UserService,
	product ProductService,
	order OrderService,
	token TokenService,
) *Services {
	return &Services{
		User:    user,
		Product: product,
		Order:   order,
		Token:   token,
	}
}
