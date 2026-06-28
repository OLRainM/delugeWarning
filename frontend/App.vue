<script>
	export default {
		onLaunch: function() {
			// 启动时恢复登录态（App 自身用 this，不能用 getApp()）
			const token = uni.getStorageSync('token');
			const user = uni.getStorageSync('user');
			if (token) {
				this.globalData.token = token;
				this.globalData.user = user;
			}
		},
		globalData: {
			// 后端基础地址，真机调试请改为可访问的域名/IP
			baseURL: 'http://127.0.0.1:8080',
			token: '',
			user: null
		}
	}
</script>

<style>
	/* 全局样式 */
	page {
		background: #f4f5f7;
		font-size: 28rpx;
		color: #1f2329;
	}
	.container { padding: 24rpx; }
	.card {
		background: #fff;
		border-radius: 16rpx;
		padding: 24rpx;
		margin-bottom: 20rpx;
		box-shadow: 0 2rpx 8rpx rgba(0,0,0,0.04);
	}
	.btn-primary { background: #1f6feb; color: #fff; border-radius: 12rpx; }
	.btn-line { border: 1rpx solid #1f6feb; color: #1f6feb; background: #fff; border-radius: 12rpx; }
	.muted { color: #8a9099; font-size: 24rpx; }
	.tag {
		display: inline-block;
		padding: 4rpx 16rpx;
		border-radius: 20rpx;
		font-size: 22rpx;
		color: #fff;
	}
	.level-blue { background: #1f6feb; }
	.level-yellow { background: #e3a008; }
	.level-orange { background: #f56300; }
	.level-red { background: #e5484d; }
	.nav-grid { display: flex; gap: 20rpx; margin-top: 8rpx; }
	.nav-item {
		flex: 1;
		background: #fff;
		border-radius: 16rpx;
		text-align: center;
		padding: 32rpx 0;
		color: #1f6feb;
	}
</style>
