Name:           ovfenv-installer
Version:        %{_version}
Release:        %{_release}
Summary:        Configure networking from vSphere ovfEnv properties

Group:          System Tools
License:        Apache 2
URL:            https://github.com/subchen/ovfenv-installer
Packager:       Guoqiang Chen <subchen@gmail.com>

BuildRoot:      %{_topdir}/BUILDROOT
Prefix:         /
AutoReqProv:    no

%description
Configure networking from vSphere ovfEnv properties

%prep

%build

%install
install -D -m0755 %{_topdir}/%{name} $RPM_BUILD_ROOT/usr/local/bin/%{name}

%clean

%files
%defattr(-,root,root)
/usr/local/bin/*

%post

%postun

%changelog
