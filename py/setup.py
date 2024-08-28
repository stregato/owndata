from setuptools import setup, find_packages, Extension

# Dummy extension to ensure the package is recognized as platform-specific
dummy_extension = Extension(
    'stash.dummy', sources=[]  # No source files, just to mark the package as platform-specific
)

# Setup script
setup(
    name='pstash',  # Generic name, if you don't need architecture-specific names
    version='0.1.1',
    packages=find_packages(),
    include_package_data=True,
    package_data={
        'stash': ['_libs/**/*'],  # Include all files under stash/_libs
    },
    ext_modules=[dummy_extension],
    author='Francesco Ink',
    author_email='me@francesco.ink',
    description='P Stash provides encrypted storage and data exchange for Python applications.',
    long_description=open('README.md').read(),
    long_description_content_type='text/markdown',
    url='https://github.com/stregato/stash',
    classifiers=[
        'Programming Language :: Python :: 3',
        'Operating System :: MacOS :: MacOS X',
        'Operating System :: POSIX :: Linux',
        'Operating System :: Microsoft :: Windows',
    ],
)