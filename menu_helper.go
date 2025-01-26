// menu_helper.go

package main

func renderMenu() string {
	return `
	<nav class="navbar navbar-expand-lg navbar-light bg-light mb-4">
		<a class="navbar-brand" href="/">🦆 Quacker</a>
		<div class="collapse navbar-collapse">
			<ul class="navbar-nav mr-auto">
				<li class="nav-item">
					<a class="nav-link" href="/sites">Sites</a>
				</li>
				<li class="nav-item">
					<a class="nav-link" href="https://github.com/mreider">GitHub</a>
				</li>
				<li class="nav-item">
					<a class="nav-link" href="https://buymeacoffee.com/mreider">Buy Me a Coffee</a>
				</li>
			</ul>
		</div>
	</nav>`
}
