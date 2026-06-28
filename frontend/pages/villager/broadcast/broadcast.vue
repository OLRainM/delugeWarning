<template>
	<view class="container">
		<view class="card">
			<view style="font-size:34rpx;font-weight:bold;">{{info.title}}</view>
			<view style="margin-top:20rpx;line-height:1.6;">{{info.content}}</view>
		</view>
		<button class="btn-primary" @tap="togglePlay">
			{{playing ? '⏸ 暂停广播' : '▶ 播放语音广播'}}
		</button>
		<view class="muted" style="text-align:center;margin-top:16rpx;">点击播放应急广播语音(TTS)</view>
	</view>
</template>

<script>
	import api from '@/utils/request';
	let audioCtx = null;
	export default {
		data() {
			return { info: {}, playing: false };
		},
		onLoad(q) {
			api.get('/api/v1/alerts/' + q.id + '/broadcast').then((info) => {
				this.info = info;
			});
		},
		methods: {
			togglePlay() {
				const url = this.info.tts_url;
				if (!url) { uni.showToast({ title: '暂无语音', icon: 'none' }); return; }
				if (!audioCtx) {
					audioCtx = uni.createInnerAudioContext();
					audioCtx.onEnded(() => { this.playing = false; });
					audioCtx.onError(() => { this.playing = false; uni.showToast({ title: '播放失败', icon: 'none' }); });
				}
				if (this.playing) {
					audioCtx.pause();
					this.playing = false;
				} else {
					audioCtx.src = url;
					audioCtx.play();
					this.playing = true;
				}
			}
		},
		onUnload() {
			if (audioCtx) { audioCtx.destroy(); audioCtx = null; }
		}
	};
</script>
