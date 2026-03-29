package service

type Services struct {
	User    *UserService
	Product *ProductService
	Order   *OrderService
	Token   *TokenService
	Cart    *CartService
	Review  *ReviewService
	Address *AddressService
}

func NewServices(
	user *UserService,
	product *ProductService,
	order *OrderService,
	token *TokenService,
	cart *CartService,
	review *ReviewService,
	address *AddressService,
) *Services {
	return &Services{
		User:    user,
		Product: product,
		Order:   order,
		Token:   token,
		Cart:    cart,
		Review:  review,
		Address: address,
	}
}
