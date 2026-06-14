// Package carbon mirrors relay-api's hard-coded carbon model (plan.md §7) so
// the engine's net-carbon gate uses the same constants.
package carbon

// kg CO2e saved vs baseline (new purchase + warehouse return + restock).
var savedByChannel = map[string]float64{
	"exchange":   1.8,
	"rescue":     2.4,
	"p2p_resale": 3.1,
	"refurb":     2.0,
	"donate":     1.5,
	"recycle":    0.6,
	"restock":    0.0,
}

// kg CO2 per km of last-mile delivery (light vehicle).
const DeliveryCO2PerKm = 0.12

// NetSaved returns channel constant minus last-mile delivery cost.
func NetSaved(channel string, deliveryKm float64) float64 {
	base := savedByChannel[channel]
	v := base - deliveryKm*DeliveryCO2PerKm
	return round3(v)
}

func round3(v float64) float64 {
	return float64(int64(v*1000+0.5*sign(v))) / 1000
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
