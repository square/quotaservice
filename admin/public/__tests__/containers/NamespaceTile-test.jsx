import { shallow } from 'enzyme';
import toJSON from 'enzyme-to-json';
import React from 'react';

import NamespaceTile from '../../src/containers/NamespaceTile.jsx';

describe('NamespaceTile', () => {
  let props, component;

  beforeEach(() => {
    props = {
      namespace: {
        name: 'namespace-name',
        dynamic_bucket_template: { name: 'dynamic_bucket_template-name' },
        default_bucket: { name: 'dynamic_bucket_template-name' },
        buckets: {
          bucket1: { name: 'bucket1-name' },
          bucket2: { name: 'bucket2-name' },
        }
      },
      selectNamespace: () => null,
      isSelected: false,
      capabilitiesEnabled: false,
      capabilities: {},
      eventEmitter: {
        dispatchEvent: jest.fn(),
      }
    };
  });

  it('renders with minimum data', () => {
    component = shallow(<NamespaceTile {...props} />);
    expect(toJSON(component)).toMatchSnapshot();
  });

  it('renderes in selected state', () => {
    props.isSelected = true;
    component = shallow(<NamespaceTile {...props} />);
    expect(toJSON(component)).toMatchSnapshot();
    expect(component.find('.namespace.selected')).toHaveLength(1);
  });

  describe('capabilities', () => {
    beforeEach(() => {
      props.capabilitiesEnabled = true;
      component = shallow(<NamespaceTile {...props} />);
    });

    it('renders ok when can not make changes', () => {
      component.setState({ canMakeChanges: false });
      expect(toJSON(component)).toMatchSnapshot();
      expect(props.eventEmitter.dispatchEvent.mock.calls).toHaveLength(1);
      expect(component.find('.canMakeChanges')).toHaveLength(0);
    });

    it('renders ok when able to make changes', () => {
      component.setState({ canMakeChanges: true });
      expect(toJSON(component)).toMatchSnapshot();
      expect(props.eventEmitter.dispatchEvent.mock.calls).toHaveLength(1);
      expect(component.find('.canMakeChanges')).toHaveLength(1);
    });
  });
});
