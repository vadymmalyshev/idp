import { h } from 'preact';
import style from './style';
import Header from '../../components/header';

const Home = () => (
	<div class={style.home}>
		<h1>Home</h1>
		<p>This is the Home component.</p>
		<Header />
	</div>
);

export default Home;
