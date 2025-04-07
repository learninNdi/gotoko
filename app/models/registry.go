package models

type Model struct {
	Model interface{}
}

func RegisterModels() []Model {
	return []Model{
		{Model: Address{}},
		{Model: User{}},
		{Model: Product{}},
		{Model: ProductImage{}},
		{Model: Section{}},
		{Model: Category{}},
		{Model: Order{}},
		{Model: OrderItem{}},
		{Model: OrderCustomer{}},
		{Model: Payment{}},
		{Model: Shipment{}},
	}
}
