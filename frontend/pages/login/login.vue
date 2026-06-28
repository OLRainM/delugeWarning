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

		<button class="btn-primary" style="margin-top:24rpx;" @tap="login">一键登录（测试模式）</button>

		<!-- 测试账户快捷入口 -->
		<view class="card" style="margin-top:24rpx;">
			<view style="font-size:24rpx;color:#8a9099;margin-bottom:16rpx;">测试账户快速登录</view>
			<view style="display:flex;gap:20rpx;">
				<view class="btn-line" style="flex:1;text-align:center;padding:20rpx;font-size:26rpx;"
					@tap="quickLogin('villager')">👤 测试村民</view>
				<view class="btn-line" style="flex:1;text-align:center;padding:20rpx;font-size:26rpx;"
					@tap="quickLogin('gridworker')">🔧 测试网格员</view>
			</view>
		</view>
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
			// 普通登录：用角色名作为 code（本地 mock 模式）
			login() { this.doLogin('user-' + this.role, this.role); },
			// 快捷测试账户：固定 code，后端会生成固定 openid
			quickLogin(role) { this.doLogin('test-' + role, role); },
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
