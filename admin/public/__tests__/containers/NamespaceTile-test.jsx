import { shallow } from 'enzyme';
import toJSON from 'enzyme-to-json';
import React from 'react';

import NamespaceTile from '../../src/containers/NamespaceTile.jsx';

describe('NamespaceTile', () => {
  it('renders with minimum data', () => {
    const props = {
      namespace: {},
      selectNamespace: () => null,
      isSelected: false,
    };
    const component = shallow(<NamespaceTile {...props} />);
    const tree = toJSON(component);
    expect(tree).toMatchSnapshot();
  });

  it('renderes in selected state', () => {
    const props = {
      namespace: {},
      selectNamespace: () => null,
      isSelected: true,
    };
    const component = shallow(<NamespaceTile {...props} />);
    const tree = toJSON(component);
    expect(tree).toMatchSnapshot();
    expect(component.find('.namespace.selected')).toHaveLength(1);
  });
});
