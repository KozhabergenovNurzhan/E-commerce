package service

type Services struct {
	User    UserServiceI
	Product ProductServiceI
	Order   OrderServiceI
	Token   TokenServiceI
	Cart    CartServiceI
	Review  ReviewServiceI
	Address AddressServiceI
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
