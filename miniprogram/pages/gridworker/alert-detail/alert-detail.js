// pages/gridworker/alert-detail/alert-detail.js —— 预警详情/复核/归档
const api = require('../../../utils/request');

Page({
  data: { id: 0, alert: {}, logs: [] },
  onLoad(q) { this.setData({ id: Number(q.id) }, () => this.load()); },
  load() {
    api.get('/api/v1/alerts/' + this.data.id).then((res) => {
      this.setData({ alert: res.alert || {}, logs: res.logs || [] });
    });
  },
  review(e) {
    const action = e.currentTarget.dataset.action;
    api.post('/api/v1/alerts/' + this.data.id + '/review', { action }).then(() => {
      wx.showToast({ title: '操作成功' });
      this.load();
    });
  },
  archive() {
    api.post('/api/v1/alerts/' + this.data.id + '/archive', {}).then(() => {
      wx.showToast({ title: '已归档' });
      this.load();
    });
  }
});
