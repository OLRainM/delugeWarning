// pages/villager/guide/guide.js —— 避险指引
const api = require('../../../utils/request');
Page({
  data: { tips: [] },
  onLoad() {
    api.get('/api/v1/guides', { disaster_type: 'flood' }).then((res) => {
      this.setData({ tips: res.tips || [] });
    });
  }
});
