package rgbw

import "MagicMirrorGo/data"
import rpio "github.com/stianeikeland/go-rpio/v4"

var w1 rpio.Pin = rpio.Pin(1)
var r1 rpio.Pin = rpio.Pin(2)
var g1 rpio.Pin = rpio.Pin(3)
var b1 rpio.Pin = rpio.Pin(4)

var w2 rpio.Pin = rpio.Pin(5)
var r2 rpio.Pin = rpio.Pin(6)
var g2 rpio.Pin = rpio.Pin(7)
var b2 rpio.Pin = rpio.Pin(8)

type LEDHelper struct{
	w1 rpio.Pin
	r1 rpio.Pin
	g1 rpio.Pin
	b1 rpio.Pin

	w2 rpio.Pin
	r2 rpio.Pin
	g2 rpio.Pin
	b2 rpio.Pin
}

func NewLEDHelper() *LEDHelper{
	// TODO get actual pins
	ret := LEDHelper{
		w1: rpio.Pin(1),
		r1: rpio.Pin(2),
		g1: rpio.Pin(3),
		b1: rpio.Pin(4),

		w2: rpio.Pin(5),
		r2: rpio.Pin(6),
		g2: rpio.Pin(7),
		b2: rpio.Pin(8),
	}
	ret.w1.Pwm()
	ret.r1.Pwm()
	ret.g1.Pwm()
	ret.b1.Pwm()
	ret.w2.Pwm()
	ret.r2.Pwm()
	ret.g2.Pwm()
	ret.b2.Pwm()
	return &ret
}

// Quite possible that I may need create a python script to control the software PWMs on the py, and have
// go act as the glue holding things together...
func (l *LEDHelper) SetLED(led data.LED){
	l.w1.Freq(led.W)
}

