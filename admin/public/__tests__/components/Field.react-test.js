import React from 'react'
import { shallow } from 'enzyme'

import Field from '../../src/components/Field.jsx'

function setup() {
  const handleChange = jest.fn()
  const props = {
    parent: "foo",
    keyName: "bar",
    handleChange: jest.fn(() => handleChange)
  }

  const enzymeWrapper = shallow(<Field {...props} />)

  return {
    handleChange,
    enzymeWrapper
  }
}

describe('Field', () => {
  it('handles integer entry', () => {
    const { handleChange, enzymeWrapper } = setup()

    enzymeWrapper.find('input').simulate('change', {
      target: { value: "123123" }
    })

    expect(enzymeWrapper.state("value")).toEqual("123123")
    expect(handleChange.mock.calls).toEqual([
      [123123]
    ])
  })

  it('handles empty entry', () => {
    const { handleChange, enzymeWrapper } = setup()

    enzymeWrapper.find('input').simulate('change', {
      target: { value: "" }
    })

    expect(enzymeWrapper.state("value")).toEqual("")
    expect(handleChange.mock.calls).toEqual([
      [null]
    ])
  })

  it('handles invalid integer entry', () => {
    const { handleChange, enzymeWrapper } = setup()

    enzymeWrapper.find('input').simulate('change', {
      target: { value: "hello" }
    })

    expect(enzymeWrapper.state("value")).toEqual("hello")
    expect(handleChange.mock.calls).toEqual([])

    enzymeWrapper.find('input').simulate('change', {
      target: { value: "-111000000000000000000000000000000000" }
    })

    expect(enzymeWrapper.state("value")).toEqual("-111000000000000000000000000000000000")
    expect(handleChange.mock.calls).toEqual([])

    enzymeWrapper.find('input').simulate('change', {
      target: { value: "111000000000000000000000000000000000" }
    })

    expect(enzymeWrapper.state("value")).toEqual("111000000000000000000000000000000000")
    expect(handleChange.mock.calls).toEqual([])
  })
})
