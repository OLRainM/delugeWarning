// utils/request.js —— 统一封装后端请求，自动带 token
const app = getApp();

function baseURL() {
  return getApp().globalData.baseURL;
}

function request(method, path, data) {
  return new Promise((resolve, reject) => {
    const token = getApp().globalData.token;
    wx.request({
      url: baseURL() + path,
      method,
      data: data || {},
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: 'Bearer ' + token } : {})
      },
      success(res) {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data);
        } else {
          wx.showToast({ title: (res.data && res.data.error) || '请求失败', icon: 'none' });
          reject(res.data);
        }
      },
      fail(err) {
        wx.showToast({ title: '网络错误', icon: 'none' });
        reject(err);
      }
    });
  });
}

const levelMap = {
  blue: { text: '蓝色', cls: 'level-blue' },
  yellow: { text: '黄色', cls: 'level-yellow' },
  orange: { text: '橙色', cls: 'level-orange' },
  red: { text: '红色', cls: 'level-red' }
};

module.exports = {
  get: (p, d) => request('GET', p + toQuery(d)),
  post: (p, d) => request('POST', p, d),
  levelMap
};

function toQuery(obj) {
  if (!obj) return '';
  const parts = Object.keys(obj)
    .filter((k) => obj[k] !== undefined && obj[k] !== '')
    .map((k) => `${k}=${encodeURIComponent(obj[k])}`);
  return parts.length ? '?' + parts.join('&') : '';
}
