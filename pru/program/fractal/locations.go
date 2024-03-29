package fractal

type Location struct {
	XCenter float64
	YCenter float64
	Zoom    float64
}

var Seeds = []Location{
	{
		XCenter: -0.722,
		YCenter: 0.246,
		Zoom:    52.631578947368425,
	},
	{
		XCenter: -0.7463,
		YCenter: 0.1102,
		Zoom:    200,
	},
	{
		XCenter: -0.7453,
		YCenter: 0.1127,
		Zoom:    1538.4615384615386,
	},
	{
		XCenter: -0.74529,
		YCenter: 0.113075,
		Zoom:    6666.666666666667,
	},
	{
		XCenter: -0.745428,
		YCenter: 0.113009,
		Zoom:    33333.333333333336,
	},
	{
		XCenter: -0.16,
		YCenter: 1.0405,
		Zoom:    38.46153846153846,
	},
	{
		XCenter: -0.925,
		YCenter: 0.266,
		Zoom:    31.25,
	},
	{
		XCenter: -1.25066,
		YCenter: 0.02012,
		Zoom:    5882.35294117647,
	},
	{
		XCenter: -0.748,
		YCenter: 0.1,
		Zoom:    714.2857142857143,
	},
	{
		XCenter: -0.235125,
		YCenter: 0.827215,
		Zoom:    24999.999999999996,
	},
	{
		XCenter: -0.722,
		YCenter: 0.246,
		Zoom:    52.631578947368425,
	},
	{
		XCenter: -1.315180982097868,
		YCenter: 0.073481649996795,
		Zoom:    10000000000000,
	},
	{
		XCenter: -0.156653458,
		YCenter: 1.039128122,
		Zoom:    499999999.99999994,
	},
	{
		XCenter: -0.1568046,
		YCenter: 1.0390207,
		Zoom:    1000000000000,
	},
	{
		XCenter: -0.16070135,
		YCenter: 1.0375665,
		Zoom:    10000000,
	},
	{
		XCenter: 0.2549870375144766,
		YCenter: -0.0005679790528465,
		Zoom:    10000000000000,
	},
	{
		XCenter: 0.267235642726,
		YCenter: -0.003347589624,
		Zoom:    8695652173.913044,
	},
	{
		XCenter: -0.0452407411,
		YCenter: 0.986816213,
		Zoom:    5714285.714285715,
	},
	{
		XCenter: -0.0452407411,
		YCenter: 0.9868162204352258,
		Zoom:    227272727.27272728,
	},
	{
		XCenter: -0.0452407411,
		YCenter: 0.9868162204352258,
		Zoom:    1470588235.2941177,
	},
	{
		XCenter: -0.0452407411,
		YCenter: 0.9868162204352258,
		Zoom:    3703703703.703704,
	},
	{
		XCenter: -0.04524074130409,
		YCenter: 0.9868162207157838,
		Zoom:    434782608695.6522,
	},
	{
		XCenter: -0.04524074130409,
		YCenter: 0.9868162207157852,
		Zoom:    14705882352941.176,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    194174757281.5534,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    800000000000,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    2105263157894.7368,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    5617977528089.888,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    11764705882352.941,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    25000000000000,
	},
	{
		XCenter: 0.281717921930775,
		YCenter: 0.5771052841488505,
		Zoom:    52083333333333.336,
	},
	{
		XCenter: -0.840719,
		YCenter: 0.22442,
		Zoom:    12658.227848101267,
	},
	{
		XCenter: -0.81153120295763,
		YCenter: 0.20142958206181,
		Zoom:    3333.3333333333335,
	},
	{
		XCenter: -0.81153120295763,
		YCenter: 0.20142958206181,
		Zoom:    169491.5254237288,
	},
	{
		XCenter: -0.81153120295763,
		YCenter: 0.20142958206181,
		Zoom:    21739130.43478261,
	},
	{
		XCenter: -0.81153120295763,
		YCenter: 0.20142958206181,
		Zoom:    662251655.6291391,
	},
	{
		XCenter: -0.81153120295763,
		YCenter: 0.20142958206181,
		Zoom:    8928571428571.428,
	},
	{
		XCenter: -0.8115312340458353,
		YCenter: 0.2014296112433656,
		Zoom:    29411764705882.35,
	},
	{
		XCenter: 0.452721018749286,
		YCenter: 0.39649427698014,
		Zoom:    9090909090909.092,
	},
	{
		XCenter: 0.45272105023,
		YCenter: 0.396494224267,
		Zoom:    370370370.3703703,
	},
	{
		XCenter: 0.45272105023,
		YCenter: 0.396494224267,
		Zoom:    2564102564.1025643,
	},
	{
		XCenter: 0.45272105023,
		YCenter: 0.396494224267,
		Zoom:    7142857142.857142,
	},
	{
		XCenter: -1.1533577030005,
		YCenter: 0.307486987838885,
		Zoom:    1886792452.8301885,
	},
	{
		XCenter: -1.1533577030005,
		YCenter: 0.307486987838885,
		Zoom:    10526315789473.684,
	},
	{
		XCenter: -1.15412664822215,
		YCenter: 0.30877492767139,
		Zoom:    322580645.1612903,
	},
	{
		XCenter: -1.15412664822215,
		YCenter: 0.30877492767139,
		Zoom:    16129032258.064514,
	},
	{
		XCenter: -1.15412664822215,
		YCenter: 0.30877492767139,
		Zoom:    105263157894.73685,
	},
	{
		XCenter: -1.15412664822215,
		YCenter: 0.30877492767139,
		Zoom:    270270270270.27026,
	},
	{
		XCenter: -1.7590170270659,
		YCenter: 0.01916067191295,
		Zoom:    909090909090.9092,
	},
	{
		XCenter: -1.99999911758738,
		YCenter: 0,
		Zoom:    675675675675.6757,
	},
	{
		XCenter: -1.99999911758738,
		YCenter: 0,
		Zoom:    1694915254237.288,
	},
	{
		XCenter: -1.99999911758738,
		YCenter: 0,
		Zoom:    4000000000000,
	},
	{
		XCenter: 0.432539867562512,
		YCenter: 0.226118675951765,
		Zoom:    312500,
	},
	{
		XCenter: 0.432539867562512,
		YCenter: 0.226118675951765,
		Zoom:    3125000000000,
	},
	{
		XCenter: 0.432539867562512,
		YCenter: 0.226118675951765,
		Zoom:    13698630136986.3,
	},
	{
		XCenter: 0.432539867562512,
		YCenter: 0.226118675951818,
		Zoom:    54945054945054.945,
	},
	{
		XCenter: 0.3369844464869,
		YCenter: 0.048778219666,
		Zoom:    55555555555.55556,
	},
	{
		XCenter: 0.3369844464873,
		YCenter: 0.0487782196791,
		Zoom:    238095238095.2381,
	},
	{
		XCenter: 0.33698444648918,
		YCenter: 0.048778219681,
		Zoom:    4761904761904.762,
	},
	{
		XCenter: 0.2929859127507,
		YCenter: 0.6117848324958,
		Zoom:    1492537.3134328357,
	},
	{
		XCenter: 0.2929859127507,
		YCenter: 0.6117848324958,
		Zoom:    1162790697.6744184,
	},
	{
		XCenter: 0.2929859127507,
		YCenter: 0.6117848324958,
		Zoom:    22727272727.272724,
	},
	{
		XCenter: 0.2929859127507,
		YCenter: 0.6117848324958,
		Zoom:    100000000000,
	},
	{
		XCenter: -0.936532336,
		YCenter: 0.2633616,
		Zoom:    5714285.714285715,
	},
	{
		XCenter: -0.7336438924199521,
		YCenter: 0.2455211406714035,
		Zoom:    2325581395.348837,
	},
	{
		XCenter: -0.7336438924199521,
		YCenter: 0.2455211406714035,
		Zoom:    22222222222222.223,
	},
}
