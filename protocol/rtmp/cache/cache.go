package cache

import (
	"github.com/chwjbn/livesrv/av"
	"github.com/chwjbn/livesrv/configure"
)

type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(configure.Config.GetInt("gop_num")),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (cache *Cache) Write(p av.Packet) {

	if p.IsMetadata {
		cache.metadata.Write(&p)
		return
	}

	if p.IsVideo {

		vh, ok := p.Header.(av.VideoPacketHeader)

		if ok {
			if vh.IsSeq() {
				cache.videoSeq.Write(&p)
			}
		}

		return
	}

	if p.IsAudio {
		ah, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			if ah.SoundFormat() == av.SOUND_AAC &&
				ah.AACPacketType() == av.AAC_SEQHDR {
				cache.audioSeq.Write(&p)
			}
		}

		return
	}

	cache.gop.Write(&p)
}

func (cache *Cache) Send(w av.WriteCloser) error {

	if err := cache.metadata.Send(w); err != nil {
		return err
	}

	if err := cache.videoSeq.Send(w); err != nil {
		return err
	}

	if err := cache.audioSeq.Send(w); err != nil {
		return err
	}

	if err := cache.gop.Send(w); err != nil {
		return err
	}

	return nil
}
