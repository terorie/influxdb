// Libraries
import React, {PureComponent} from 'react'
import _ from 'lodash'

// Components
import NavMenuItem from 'src/page_layout/components/NavMenuItem'

import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  navItems: NavItem[]
}

interface NavItem {
  title: string
  link: string
  icon: string
  location: string
  highlightWhen: string[]
}

@ErrorHandling
class NavMenu extends PureComponent<Props> {
  constructor(props) {
    super(props)
  }

  public render() {
    const {navItems} = this.props

    return (
      <nav className="nav">
        {navItems.map(({title, highlightWhen, icon, link, location}) => (
          <NavMenuItem
            key={`navigation--${title}`}
            title={title}
            highlightWhen={highlightWhen}
            link={link}
            location={location}
          >
            <span className={`icon sidebar--icon ${icon}`} />
          </NavMenuItem>
        ))}
      </nav>
    )
  }
}

export default NavMenu
