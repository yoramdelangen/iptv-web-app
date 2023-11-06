import 'video.js/dist/video-js.min.css';
import './style.css';

import videojs from 'video.js';
// import typescriptLogo from './typescript.svg'
// import viteLogo from '/vite.svg'
// import { setupCounter } from './counter.ts'

const STREAM_ID = 'http://localhost:3000/stream/123456/movies/226947.mkv';
const CONTAINER_FULL_PAGE_CLASS = 'full-page'

window.startPlaying = function(type: string, src: string) {
  console.log(player)
  player.src({ type, src });
  // player.play();
}


// document.querySelector<HTMLDivElement>('#app')!.innerHTML += `
//   <div>
//     Showing the root project
//   </div>
// `
//
// setupCounter(document.querySelector<HTMLButtonElement>('#counter')!)

const container = document.querySelector('.vid-container');
const playBtn = container?.querySelector('.action.play')
const pauseBtn = container?.querySelector('.action.pause')

const player = videojs('vid_player', {
  controls: true,
  autoplay: false,
  preload: 'auto'
});
player.ready(function() {
  player.play();
});
player.on('pause', () => {
  pauseBtn?.classList.add('!hidden');
  playBtn?.classList.remove('!hidden');
  console.log('pause video')
})
player.on('play', () => {
  console.log('playing video')
  playBtn?.classList.add('!hidden');
  pauseBtn?.classList.remove('!hidden');
})


const temp = document.getElementById('temp')
temp?.addEventListener('click', () => {
  window.startPlaying('video/webm', STREAM_ID)
})

const minimize = container?.querySelector('.minimize');
minimize?.addEventListener('click', () => {
  if (container?.classList.contains(CONTAINER_FULL_PAGE_CLASS)) {
    container?.classList.remove(CONTAINER_FULL_PAGE_CLASS);
  } else {
    container?.classList.add(CONTAINER_FULL_PAGE_CLASS);
  }
});

const titleBar = container?.querySelector('.title-bar')
titleBar?.addEventListener('click', (e: Event) => {
  e.preventDefault();
  container?.classList.add(CONTAINER_FULL_PAGE_CLASS);
})

playBtn?.addEventListener('click', (e: Event) => {
  e.preventDefault()
  player.play();
})
pauseBtn?.addEventListener('click', (e: Event) => {
  e.preventDefault()
  player.pause();
})
