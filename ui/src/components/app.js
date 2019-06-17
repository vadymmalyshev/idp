import { h, Component } from 'preact';
import { Router, route } from 'preact-router';
import fetch from 'unfetch';
import Header from './header';
import API from '../config/api';
// Code-splitting is automated for routes
import Home from '../routes/home';
import Profile from '../routes/profile';
import Login from '../routes/login';
import Forgot from '../routes/forgot';
import Register from '../routes/register';

export default class App extends Component {
	
	/** Gets fired when the route changes.
	 *	@param {Object} event		"change" event from [preact-router](http://git.io/preact-router)
	 *	@param {string} event.url	The newly routed URL
	 */
	
	// some method that returns a promise
	isAuthenticated = () => {
		return false;
		return fetch(API.account, {
      method: 'POST',
      // body: JSON.stringify({
      //   email: login, password
      // }),
      // mode: 'cors',
      // credentials: 'include',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        // 'Origin': 'https://id.hiveon.dev',
        // 'Refferer': 'https://id.hiveon.dev/login',
        // 'Host': 'https://id.hiveon.dev',
        // ...(cookie ? { Cookie: cookie } : null),
      },
		});
		// var promise1 = new Promise(function(resolve, reject) {
		// 	setTimeout(function() {
		// 		resolve('foo');
		// 	}, 300);
		// });
		// return promise1
	}
	
  handleRoute = async e => {
		this.currentUrl = e.url;
    switch (e.url) {
      case '/profile':
				const isAuthed = await this.isAuthenticated();
				// debugger;
				if (!isAuthed) route('/login', true);
				this.currentUrl = e.url;
				break;
			default: 
				this.currentUrl = e.url;
    }
  };

	render() {
		return (
			<div id="app">
				{/* <Header /> */}
				<Router onChange={this.handleRoute}>
					<Home path="/" />
					<Login path="/login" />
					<Forgot path="/restore-pass" />
					<Register path="/register" />
					<Profile path="/profile/" user="me" />
					<Profile path="/profile/:user" />
				</Router>
			</div>
		);
	}
}
