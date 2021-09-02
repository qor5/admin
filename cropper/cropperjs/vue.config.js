module.exports = {
	runtimeCompiler: true,
	productionSourceMap: false,
	devServer: {
		port: 3600
	},
	configureWebpack: {
		output: {
			libraryExport: 'default'
		},
		externals: { vue: "Vue" },
	}
}
