package service

type Services struct {
	User    UserService
	Product ProductService
	Order   OrderService
	Token   TokenService
	Cart    CartService
}

func NewServices(
	user UserService,
	product ProductService,
	order OrderService,
	token TokenService,
	cart CartService,
) *Services {
	return &Services{
		User:    user,
		Product: product,
		Order:   order,
		Token:   token,
		Cart:    cart,
	}
}
