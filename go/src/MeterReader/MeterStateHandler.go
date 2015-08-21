package MeterReader

import (
	"log"
)

type Meter struct {
	MeterId       int32
	Name          string
	Unit          string
	CurrentSeries uint32
	StartCount    uint64
	LastCount     uint64
}

type MeterUpdate struct {
	MeterId int32  `json:"meter"`
	Unit    string `json:"unit"`
	Name    string `json:"name"`
	Value   uint64 `json:"value"`
}

type MeterStateHandler struct {
	mdb    *MeterDB
	meters map[int32]*Meter
}

func NewMeterStateHandler() *MeterStateHandler {
	msh := new(MeterStateHandler)
	msh.mdb = NewMeterDB()
	msh.meters = msh.mdb.GetMeterState()

	return msh
}

func (msh *MeterStateHandler) Handle(queue chan *CounterUpdate) chan *MeterUpdate {
	outch := make(chan *MeterUpdate)

	go func() {
		for msg := range queue {
			tmsg := msh.Translate(msg)
			outch <- tmsg
		}
	}()

	return outch
}

func (msh *MeterStateHandler) Translate(msg *CounterUpdate) *MeterUpdate {

	//Retreive client information from the protobuf message
	MeterId := msg.GetMeterId()

	meter, ok := msh.meters[MeterId]
	if !ok {
		meter = &Meter{MeterId: MeterId}
		msh.meters[MeterId] = meter
		log.Printf("Creating new meter instance id %s\n", MeterId)
	}

	SeriesId := msg.GetSeriesId()
	CurrentCounterValue := msg.GetCurrentCounterValue()

	if meter.CurrentSeries != SeriesId {
		meter.CurrentSeries = SeriesId
		meter.StartCount = meter.LastCount
	}
	meter.LastCount = meter.StartCount + CurrentCounterValue

	log.Printf("meterid=%d series=%d counter=%d -> absolute=%d\n", MeterId, SeriesId, CurrentCounterValue, meter.LastCount)

	umsg := MeterUpdate{MeterId: meter.MeterId, Unit: meter.Unit, Name: meter.Name, Value: meter.LastCount}

	return &umsg
}
