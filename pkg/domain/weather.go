package domain

type Weather struct {
	Location       string
	Temp           float64 // Celsius
	TempFeel       float64 // Celsius
	Pressure       int     // hPa
	Humidity       int     // %
	Weather        string  // for icon
	WeatherVerbose string
	WindSpeed      float64 // meter/sec
	WindDirection  string
}
