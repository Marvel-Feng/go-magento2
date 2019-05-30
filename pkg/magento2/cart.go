package magento2

import (
	"fmt"
	"github.com/hermsi1337/go-magento2/internal/utils"
	"strconv"
)

type Cart struct {
	Route     string
	QuoteID   string
	Detailed  DetailedCart
	ApiClient *ApiClient
}

func (cart *Cart) GetDetails() (DetailedCart, error) {
	var detailedCart = &DetailedCart{}
	httpClient := cart.ApiClient.HttpClient
	resp, err := httpClient.R().SetResult(detailedCart).Get(cart.Route)
	if err != nil {
		return *detailedCart, fmt.Errorf("error while getting detailed cart object from magento2-api: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *detailedCart, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}
	detailedCart = resp.Result().(*DetailedCart)

	return *detailedCart, err
}

func (cart *Cart) UpdateSelf() error {
	newDetails, err := cart.GetDetails()
	if err != nil {
		return fmt.Errorf("error while updating the cart object from magento2-api: '%v'", err)
	}

	cart.Detailed = newDetails
	return nil
}

func (cart *Cart) AddItems(items []Item) error {
	endpoint := cart.Route + cartItems
	httpClient := cart.ApiClient.HttpClient

	type PayLoad struct {
		CartItem Item `json:"cartItem"`
	}

	for _, item := range items {
		item.QuoteID = cart.QuoteID
		payLoad := &PayLoad{
			CartItem: item,
		}
		resp, err := httpClient.R().SetBody(payLoad).Post(endpoint)
		if err != nil {
			return fmt.Errorf("received error while adding item '%v' to cart: '%v'", item, err)
		} else if resp.StatusCode() >= 400 {
			return fmt.Errorf("unexpected statuscode '%v' - response: '%v' - body: '%+v'", resp.StatusCode(), resp, resp.Request.Body)
		}
	}

	err := cart.UpdateSelf()
	if err != nil {
		return err
	}

	return nil
}

func (cart *Cart) EstimateShippingCarrier(addrInfo Address) ([]Carrier, error) {
	endpoint := cart.Route + cartShippingCosts
	httpClient := cart.ApiClient.HttpClient

	type PayLoad struct {
		Address Address `json:"address"`
	}

	payLoad := &PayLoad{
		Address: addrInfo,
	}

	shippingCosts := &[]Carrier{}

	resp, err := httpClient.R().SetBody(*payLoad).SetResult(shippingCosts).Post(endpoint)
	if err != nil {
		return *shippingCosts, fmt.Errorf("received erro while estimating shipping costs: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *shippingCosts, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	shippingCosts = resp.Result().(*[]Carrier)

	if len(*shippingCosts) == 0 {
		return *shippingCosts, fmt.Errorf("received no suitable shipping - response: '%v'", resp)
	}

	return *shippingCosts, nil
}

func (cart *Cart) AddShippingInformation(addrInfo AddressInformation) error {
	endpoint := cart.Route + cartShippingInformation
	httpClient := cart.ApiClient.HttpClient

	type PayLoad struct {
		AddressInformation AddressInformation `json:"addressInformation"`
	}

	payLoad := &PayLoad{
		AddressInformation: addrInfo,
	}

	resp, err := httpClient.R().SetBody(*payLoad).Post(endpoint)
	if err != nil {
		return fmt.Errorf("received error while adding shipping information: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	err = cart.UpdateSelf()
	if err != nil {
		return err
	}

	return nil
}

func (cart *Cart) EstimatePaymentMethods() ([]PaymentMethod, error) {
	endpoint := cart.Route + cartPaymentMethods
	httpClient := cart.ApiClient.HttpClient

	paymentMethods := &[]PaymentMethod{}

	resp, err := httpClient.R().SetResult(paymentMethods).Get(endpoint)
	if err != nil {
		return *paymentMethods, fmt.Errorf("received error while estimating payment methods costs: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *paymentMethods, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	paymentMethods = resp.Result().(*[]PaymentMethod)

	if len(*paymentMethods) == 0 {
		return *paymentMethods, fmt.Errorf("received no suitable payment method - response: '%v'", resp)
	}

	return *paymentMethods, nil
}

func (cart *Cart) CreateOrder(paymentMethod PaymentMethod) (*Order, error) {
	endpoint := cart.Route + cartPlaceOrder
	httpClient := cart.ApiClient.HttpClient

	type PayLoad struct {
		PaymentMethod PaymentMethodCode `json:"paymentMethod"`
	}

	payLoad := &PayLoad{
		PaymentMethod: PaymentMethodCode{
			Method: paymentMethod.Code,
		},
	}

	resp, err := httpClient.R().SetBody(payLoad).Put(endpoint)
	if err != nil {
		return nil, fmt.Errorf("received error while creating order: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	orderIDString := utils.MayTrimSurroundingQuotes(resp.String())
	orderIDInt, err := strconv.Atoi(orderIDString)
	if err != nil {
		return nil, fmt.Errorf("unexpected error while extracting orderID: '%v'", err)
	}

	return &Order{
		ID:        orderIDInt,
		ApiClient: cart.ApiClient,
		Route:     order + "/" + orderIDString,
	}, nil
}

func (cart *Cart) DeleteItem(itemID int) error {
	endpoint := cart.Route + cartItems + "/" + strconv.Itoa(itemID)
	httpClient := cart.ApiClient.HttpClient

	resp, err := httpClient.R().Delete(endpoint)
	if err != nil {
		return fmt.Errorf("received error while creating order: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	return nil
}

func (cart *Cart) DeleteAllItems() error {
	err := cart.UpdateSelf()
	if err != nil {
		return err
	}

	for i := range cart.Detailed.Items {
		err = cart.DeleteItem(cart.Detailed.Items[i].ItemID)
		if err != nil {
			return err
		}
	}

	return nil
}