import { h, Component } from 'preact';
import cx from 'classnames';
import style from './style.css';

class FormInput extends Component {

	state = {
    errors: {},
    touched: false
	};

	componentDidMount() {
	}

	componentWillUnmount() {
    this.setState({touched: false})
  }

  handleBlur = () => {
    this.setState({touched: true})
  };

	// Note: `user` comes from the URL, courtesy of our router
	render({ value, name, placeholder, className, error, ...props }, { touched }) {
		return (
          <div className={style.inputWrapper}>
          <input
            id={name}
            class={cx(style.input, {
              [className]: !!className, 
              [style.value]: !!value, 
              [style.error]: error})
            }
            name={name}
            placeholder=""
            value={value}
            onBlur={this.handleBlur}
            {...props}
          />
          <label className={style.label} for={name} htmlFor={name}>{placeholder}</label>
          <div className={style.errorMessage}>
            {error && touched
              ? error
              : ''}
          </div>
        </div>
		);
	}
}

export default FormInput;
