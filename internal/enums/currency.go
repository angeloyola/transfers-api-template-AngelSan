package enums

type Currency int

const (
	CurrencyUSD Currency = iota
	CurrencyEUR
	CurrencyGBP
	CurrencyBRL
	CurrencyARS
	CurrencyCLP
	CurrencyCOP
	CurrencyMXN
	CurrencyUnknown
)

func (c Currency) String() string {
	switch c {
	case CurrencyUSD:
		return "USD"
	case CurrencyEUR:
		return "EUR"
	case CurrencyGBP:
		return "GBP"
	case CurrencyBRL:
		return "BRL"
	case CurrencyARS:
		return "ARS"
	case CurrencyCLP:
		return "CLP"
	case CurrencyCOP:
		return "COP"
	case CurrencyMXN:
		return "MXN"
	default:
		return "unknown"
	}
}

func ParseCurrency(value string) Currency {
	switch value {
	case "USD":
		return CurrencyUSD
	case "EUR":
		return CurrencyEUR
	case "GBP":
		return CurrencyGBP
	case "BRL":
		return CurrencyBRL
	case "ARS":
		return CurrencyARS
	case "CLP":
		return CurrencyCLP
	case "COP":
		return CurrencyCOP
	case "MXN":
		return CurrencyMXN
	default:
		return CurrencyUnknown
	}
}
