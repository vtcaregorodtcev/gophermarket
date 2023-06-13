package helpers

func IsOrderNumberValid(orderNumber string) bool {
	sum := 0
	isEven := false

	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit := int(orderNumber[i] - '0')

		if isEven {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}
