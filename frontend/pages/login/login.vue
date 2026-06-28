<template>
	<view class="container">
		<view class="card" style="text-align:center;padding:48rpx 24rpx;">
			<view style="font-size:40rpx;font-weight:bold;">乡村应急广播与预警</view>
			<view class="muted" style="margin-top:12rpx;">暴雨夜应急预警联动平台</view>
		</view>

		<view class="card">
			<view style="margin-bottom:20rpx;font-weight:bold;">选择身份</view>
			<view style="display:flex;gap:20rpx;">
				<view :class="role==='villager'?'btn-primary':'btn-line'"
					style="flex:1;text-align:center;padding:24rpx;"
					@tap="chooseRole('villager')">村民</view>
				<view :class="role==='gridworker'?'btn-primary':'btn-line'"
					style="flex:1;text-align:center;padding:24rpx;"
					@tap="chooseRole('gridworker')">网格员</view>
			</view>
		</view>

		<button class="btn-primary" style="margin-top:24rpx;" @tap="login">微信一键登录</button>
		<view class="muted" style="margin-top:16rpx;text-align:center;">网格员兼任后台复核与设备管理</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { role: 'villager' };
		},
		methods: {
			chooseRole(r) { this.role = r; },
			login() {
				const role = this.role;
				// #ifdef MP-WEIXIN
				uni.login({
					success: (res) => this.doLogin(res.code, role)
				});
				// #endif
				// #ifndef MP-WEIXIN
				// 非微信端用占位 code 便于 H5/App 调试
				this.doLogin('debug-' + role, role);
				// #endif
			},
			doLogin(code, role) {
				api.post('/api/v1/auth/wx-login', { code, role }).then((data) => {
					const app = getApp();
					app.globalData.token = data.token;
					app.globalData.user = data.user;
					uni.setStorageSync('token', data.token);
					uni.setStorageSync('user', data.user);
					if (data.user.role === 'gridworker') {
						uni.reLaunch({ url: '/pages/gridworker/tasks/tasks' });
					} else {
						uni.reLaunch({ url: '/pages/villager/home/home' });
					}
				});
			}
		}
	};
</script>
