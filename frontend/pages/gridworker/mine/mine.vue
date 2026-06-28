<template>
	<view class="container">
		<view class="card">
			<view style="font-size:32rpx;font-weight:bold;">{{user.name || '网格员'}}</view>
			<view class="muted" style="margin-top:8rpx;">角色：网格员（兼后台管理与复核）</view>
			<view class="muted">所属网格：{{user.grid_id}}</view>
		</view>
		<button class="btn-line" @tap="logout">退出登录</button>
	</view>
</template>

<script>
	export default {
		data() {
			return { user: {} };
		},
		onShow() { this.user = getApp().globalData.user || {}; },
		methods: {
			logout() {
				uni.removeStorageSync('token');
				uni.removeStorageSync('user');
				getApp().globalData.token = '';
				getApp().globalData.user = null;
				uni.reLaunch({ url: '/pages/login/login' });
			}
		}
	};
</script>
