// Package ui owns panels, menus, the HUD and input mapping. It turns mouse and
// keyboard input into sim commands and hands them to a sink interface that app
// connects to the session; it never sends them over the network itself.
//
// It may import Ebitengine, sim, balance, assets and render. It must not
// import netplay or app.
package ui
