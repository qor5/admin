package content

import (
	. "github.com/theplant/docgo"
)

var Home = Doc(
	Markdown(`
QOR5 is a powerful Go library that empowers developers to craft web applications with exceptional ease and extensive customization potential. Leveraging Go's robust static typing and minimizing the need for JavaScript or TypeScript, QOR5 delivers a highly efficient development workflow while fostering a culture of component reuse.

## Revolutionizing HTML Rendering

QOR5 challenges conventional wisdom by advocating against the use of template languages for HTML generation. 
It instead invites developers to harness the expressive power of [Go's static typing to author HTML](/advanced-functions/the-go-html-builder.html) directly within
the codebase. This innovative approach confers several key advantages:

- **Enhanced Readability and Maintainability**: By embracing Go's statically typed syntax, 
projects maintain a cohesive coding style, resulting in cleaner, more comprehensible code that is effortless to navigate and update.

- **Proactive Error Detection**: The static typing inherent to Go enables rigorous compile-time checks, 
catching potential issues early in the development cycle and preventing them from surfacing as production bugs.

- **Unmatched Reusability**: QOR5 champions the modular nature of components, 
which can be seamlessly abstracted and repurposed across various aspects of an application. As these components are written in pure Go, incorporating third-party libraries becomes a breeze—simply import and utilize them like any standard Go package.

- **Streamlined Development**: By minimizing the need for JavaScript or TypeScript, 
QOR5 simplifies the development process and significantly reduces the intricacies involved in creating dynamic, interactive web experiences. Developers can concentrate on delivering core functionality without being bogged down by the complexities of multiple scripting languages.

## A Unified Development Experience

QOR5's innovative approach to HTML rendering through Go's static typing eliminates the learning curve and maintenance overhead associated with managing multiple template languages. As a result, developers enjoy a coherent, streamlined development journey that prioritizes their focus on the essence of their web applications.


## Document Structure

This comprehensive guide is structured to facilitate a hands-on understanding of QOR5's capabilities and practical application. It is anchored around a sample project, which serves as a tangible reference point throughout the learning process.

- **Quick Sample Project**: Our journey commences with a concise overview of a representative sample project, visually demonstrating QOR5's versatility and feature set.

- **Basic Functions**: This section delves into the fundamental building blocks of QOR5, 
walking you through essential features ranging from list pages to edit interfaces. It equips you with the knowledge needed to implement common administrative functionalities found in modern web applications.

- QOR5 Essentials and Advanced Functions: Here, we delve beneath the surface, examining the inner workings of QOR5. Topics include sophisticated page rendering techniques and cutting-edge features such as partial page refreshing, providing you with a deeper understanding of QOR5's full potential.

- **Digging Deeper**: In the culminating segment, you will learn how to unleash your creativity by designing and implementing custom components for QOR5. This empowering section teaches you how to extend QOR5's functionality and tailor it precisely to the unique requirements of your projects.


**Join the Discord community**: https://discord.gg/76YPsVBE4E
`)).Title("Introduction").
	Slug("/")
