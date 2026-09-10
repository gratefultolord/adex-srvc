package auction

func FilterPartners(input Input, partners []Partner) []Partner {
	filtered := make([]Partner, 0, len(partners))

	for _, partner := range partners {
		if MatchPartner(input, partner) {
			filtered = append(filtered, partner)
		}
	}

	return filtered
}

func MatchPartner(input Input, partner Partner) bool {
	if !partner.IsEnabled {
		return false
	}

	if !matchesCountry(input.Country, partner.Countries) {
		return false
	}

	if !matchesDeviceType(input.DeviceType, partner.DeviceTypes) {
		return false
	}

	if input.BidFloor < partner.MinBidFloor {
		return false
	}

	if hasBlockedCategory(
		input.Categories,
		partner.BlockedCategories,
	) {
		return false
	}

	return true
}

func matchesCountry(country string, countries []string) bool {
	if len(countries) == 0 {
		return true
	}

	for _, allowedCountry := range countries {
		if country == allowedCountry {
			return true
		}
	}

	return false
}

func matchesDeviceType(
	deviceType string,
	deviceTypes []string,
) bool {
	if len(deviceTypes) == 0 {
		return true
	}

	for _, allowedDeviceType := range deviceTypes {
		if deviceType == allowedDeviceType {
			return true
		}
	}

	return false
}

func hasBlockedCategory(
	categories []string,
	blockedCategories []string,
) bool {
	for _, category := range categories {
		for _, blockedCategory := range blockedCategories {
			if category == blockedCategory {
				return true
			}
		}
	}

	return false
}
