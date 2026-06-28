<template>
	<view class="container">
		<view class="card">
			<!-- 姓名：点击进入编辑 -->
			<view v-if="!editing">
				<view style="font-size:32rpx;font-weight:bold;">{{user.name || '（未填写姓名）'}}</view>
				<view class="muted" style="margin-top:8rpx;">角色：网格员（兼后台管理与复核）</view>
				<view class="muted">所属网格编号：{{user.grid_id}}</view>
				<view style="margin-top:20rpx;color:#1f6feb;" @tap="startEdit">✏️ 修改姓名</view>
			</view>
			<!-- 编辑态 -->
			<view v-else>
				<view style="font-weight:bold;margin-bottom:12rpx;">修改姓名</view>
				<input style="border:1rpx solid #e0e0e0;border-radius:8rpx;padding:16rpx;width:100%;box-sizing:border-box;"
					:value="nameInput" @input="e => nameInput = e.detail.value" placeholder="请输入真实姓名" />
				<view style="display:flex;gap:20rpx;margin-top:20rpx;">
					<view class="btn-primary" style="flex:1;text-align:center;padding:20rpx;" @tap="saveName">保存</view>
					<view class="btn-line" style="flex:1;text-align:center;padding:20rpx;" @tap="editing=false">取消</view>
				</view>
			</view>
		</view>
		<button class="btn-line" style="margin-top:8rpx;" @tap="logout">退出登录</button>
	</view>
</template>

<script>
	import api from '@/utils/request';
	export default {
		data() {
			return { user: {}, editing: false, nameInput: '' };
		},
		onShow() {
			this.user = getApp().globalData.user || {};
			this.nameInput = this.user.name || '';
		},
		methods: {
			startEdit() { this.nameInput = this.user.name || ''; this.editing = true; },
			saveName() {
				if (!this.nameInput.trim()) { uni.showToast({ title: '姓名不能为空', icon: 'none' }); return; }
				const app = getApp();
				uni.request({
					url: app.globalData.baseURL + '/api/v1/profile/name',
					method: 'PUT',
					header: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + app.globalData.token },
					data: { name: this.nameInput.trim() },
					success: (res) => {
						if (res.statusCode === 200) {
							this.user.name = this.nameInput.trim();
							app.globalData.user = { ...app.globalData.user, name: this.user.name };
							uni.setStorageSync('user', app.globalData.user);
							this.editing = false;
							uni.showToast({ title: '保存成功' });
						} else {
							uni.showToast({ title: '保存失败', icon: 'none' });
						}
					}
				});
			},
			logout() {
				uni.removeStorageSync('token'); uni.removeStorageSync('user');
				getApp().globalData.token = ''; getApp().globalData.user = null;
				uni.reLaunch({ url: '/pages/login/login' });
			}
		}
	};
</script>
