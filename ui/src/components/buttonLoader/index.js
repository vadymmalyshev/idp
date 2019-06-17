import { h, Component } from 'preact';
import cx from 'classnames';
import style from './style.css';

class ButtonLoader extends Component {
  static defaultProps = {
    isLoading: false,
    red: false,
    orange: false,
    blue: false,
    children: '',
    backgroundColor: null,
  };

  render = () => {
    const { isLoading, children, backgroundColor, red, orange, blue, green } = this.props;
    const shareItemProps = {
      ...(backgroundColor ? { style: { backgroundColor } } : {}),
      className: cx(style.loaderItem, {
        [style.red]: red,
        [style.orange]: orange,
        [style.blue]: blue,
        [style.green]: green,
      }),
    };

    return (
      <div className={style.fullWidthContainer}>
        {isLoading && (
          <div className={style.loader}>
            <div {...shareItemProps} />
            <div {...shareItemProps} />
            <div {...shareItemProps} />
          </div>
        )}

        <div
          className={style.fullWidthContainer}
          style={{ visibility: isLoading ? 'hidden' : 'initial' }}
        >
          {children}
        </div>
      </div>
    );
  };
}

export default ButtonLoader;
