import * as MutableActions from '../../src/actions/mutable.jsx';
import { SELECT_NAMESPACE } from '../../src/actions/namespaces.jsx';
import { selectedNamespace } from '../../src/reducers/namespaces.jsx';

describe('selectedNamespace reducer', () => {
  it('should handle initial state', () => {
    expect(selectedNamespace(undefined, {})).toEqual(null)
  })

  it('should handle ADD_NAMESPACE', () => {
    expect(selectedNamespace(undefined, {
      type: MutableActions.ADD_NAMESPACE,
      namespace: 'foo'
    })).toEqual('foo')
  })

  it('should handle REMOVE_NAMESPACE', () => {
    expect(selectedNamespace('foo', {
      type: MutableActions.REMOVE_NAMESPACE
    })).toEqual(null)
  })

  it('should handle SELECT_NAMESPACE', () => {
    expect(selectedNamespace(undefined, {
      type: SELECT_NAMESPACE,
      namespace: 'foo'
    })).toEqual('foo')
  })
})
