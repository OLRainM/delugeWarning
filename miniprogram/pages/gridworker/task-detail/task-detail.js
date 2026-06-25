// pages/gridworker/task-detail/task-detail.js —— 任务确认与处置
const api = require('../../../utils/request');

Page({
  data: { id: 0, remark: '', attachments: [] },
  onLoad(q) { this.setData({ id: Number(q.id) }); },
  confirm() {
    api.post('/api/v1/tasks/' + this.data.id + '/confirm', {}).then(() => {
      wx.showToast({ title: '已确认接收' });
    });
  },
  onRemark(e) { this.setData({ remark: e.detail.value }); },
  chooseMedia() {
    wx.chooseMedia({
      count: 3, mediaType: ['image', 'video'],
      success: (res) => {
        // 申请预签名地址并直传（mock 模式下直接记录 access_url）
        res.tempFiles.forEach((f) => this.uploadOne(f));
      }
    });
  },
  uploadOne(file) {
    const key = 'task/' + this.data.id + '/' + Date.now() + '.dat';
    api.post('/api/v1/uploads/presign', { key }).then((r) => {
      // 真机：wx.uploadFile 到 r.upload_url；此处直接登记访问地址
      const type = file.fileType === 'video' ? 'video' : 'image';
      const list = this.data.attachments.concat([{ type, cos_key: key, url: r.access_url }]);
      this.setData({ attachments: list });
      wx.showToast({ title: '已添加附件' });
    });
  },
  submit() {
    api.post('/api/v1/tasks/' + this.data.id + '/handle', {
      remark: this.data.remark, attachments: this.data.attachments
    }).then(() => {
      wx.showToast({ title: '处置已提交' });
      setTimeout(() => wx.navigateBack(), 800);
    });
  }
});
