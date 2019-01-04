import Enzyme from 'enzyme';
import Adapter from 'enzyme-adapter-react-16';

require('isomorphic-fetch');

Enzyme.configure({ adapter: new Adapter() });
