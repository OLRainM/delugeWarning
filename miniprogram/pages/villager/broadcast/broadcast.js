// pages/villager/broadcast/broadcast.js —— 广播详情与 TTS 播放
const api = require('../../../utils/request');
let audioCtx = null;

Page({
  data: { info: {}, playing: false },
  onLoad(q) {
    api.get('/api/v1/alerts/' + q.id + '/broadcast').then((info) => {
      this.setData({ info });
    });
  },
  togglePlay() {
    const url = this.data.info.tts_url;
    if (!url) { wx.showToast({ title: '暂无语音', icon: 'none' }); return; }
    if (!audioCtx) {
      audioCtx = wx.createInnerAudioContext();
      audioCtx.onEnded(() => this.setData({ playing: false }));
      audioCtx.onError(() => { this.setData({ playing: false }); wx.showToast({ title: '播放失败', icon: 'none' }); });
    }
    if (this.data.playing) {
      audioCtx.pause();
      this.setData({ playing: false });
    } else {
      audioCtx.src = url;
      audioCtx.play();
      this.setData({ playing: true });
    }
  },
  onUnload() { if (audioCtx) { audioCtx.destroy(); audioCtx = null; } }
});
